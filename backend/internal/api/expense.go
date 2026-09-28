package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/expense"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const expenseDateLayout = "2006-01-02"

type listExpensesResponse struct {
	Expenses []expense.Expense `json:"expenses"`
}

type expenseItemRequest struct {
	Description string           `json:"description"`
	CategoryID  uuid.UUID        `json:"category_id"`
	Price       *decimal.Decimal `json:"price"`
	Quantity    *decimal.Decimal `json:"quantity"`
	Discount    *decimal.Decimal `json:"discount"`
	PeriodFrom  *string          `json:"period_from"`
	PeriodTo    *string          `json:"period_to"`
}

type saveExpenseRequest struct {
	OccurredOn  string               `json:"occurred_on"`
	Currency    string               `json:"currency"`
	PaymentType *string              `json:"payment_type"`
	Note        *string              `json:"note"`
	Items       []expenseItemRequest `json:"items"`
}

type listSuggestionsResponse struct {
	Suggestions []expense.Suggestion `json:"suggestions"`
}

func (req saveExpenseRequest) toExpense() (expense.Expense, error) {
	occurredOn, err := time.Parse(expenseDateLayout, req.OccurredOn)
	if err != nil {
		return expense.Expense{}, errors.New("occurred_on must be YYYY-MM-DD")
	}

	e := expense.Expense{
		OccurredOn:  occurredOn,
		Currency:    req.Currency,
		PaymentType: req.PaymentType,
		Note:        req.Note,
		Items:       make([]expense.Item, 0, len(req.Items)),
	}

	for i, it := range req.Items {
		line := i + 1

		if it.Price == nil {
			return expense.Expense{}, fmt.Errorf("line %d: price is required", line)
		}

		quantity := decimal.NewFromInt(1)
		if it.Quantity != nil {
			quantity = *it.Quantity
		}

		discount := decimal.Zero
		if it.Discount != nil {
			discount = *it.Discount
		}

		periodFrom, err := parseOptionalDate(it.PeriodFrom)
		if err != nil {
			return expense.Expense{}, fmt.Errorf("line %d: period_from must be YYYY-MM-DD", line)
		}
		periodTo, err := parseOptionalDate(it.PeriodTo)
		if err != nil {
			return expense.Expense{}, fmt.Errorf("line %d: period_to must be YYYY-MM-DD", line)
		}

		e.Items = append(e.Items, expense.Item{
			Description: it.Description,
			CategoryID:  it.CategoryID,
			Price:       *it.Price,
			Quantity:    quantity,
			Discount:    discount,
			PeriodFrom:  periodFrom,
			PeriodTo:    periodTo,
		})
	}

	return e, nil
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(expenseDateLayout, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func writeExpenseError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, expense.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, expense.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (s *Server) handleListExpenses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		from, err := parseOptionalDate(&fromStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "from must be YYYY-MM-DD")
			return
		}
		to, err := parseOptionalDate(&toStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "to must be YYYY-MM-DD")
			return
		}

		list, err := s.expenses.List(r.Context(), from, to)
		if err != nil {
			writeExpenseError(w, err, "list expenses")
			return
		}

		writeJSON(w, http.StatusOK, listExpensesResponse{Expenses: list})
	}
}

func (s *Server) handleGetExpense() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		result, err := s.expenses.GetByID(r.Context(), id)
		if err != nil {
			writeExpenseError(w, err, "get expense")
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleCreateExpense() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req saveExpenseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		e, err := req.toExpense()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		created, err := s.expenses.Create(r.Context(), e)
		if err != nil {
			writeExpenseError(w, err, "create expense")
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateExpense() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req saveExpenseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		e, err := req.toExpense()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		e.ID = id

		updated, err := s.expenses.Update(r.Context(), e)
		if err != nil {
			writeExpenseError(w, err, "update expense")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteExpense() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.expenses.Delete(r.Context(), id); err != nil {
			writeExpenseError(w, err, "delete expense")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleListExpenseSuggestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.expenses.Suggestions(r.Context())
		if err != nil {
			writeExpenseError(w, err, "list expense suggestions")
			return
		}

		writeJSON(w, http.StatusOK, listSuggestionsResponse{Suggestions: list})
	}
}
