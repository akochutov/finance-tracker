package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/akochutov/finance-tracker/internal/expensecategory"
	"github.com/google/uuid"
)

type expenseGroupResponse struct {
	expensecategory.Group
	Categories []expensecategory.Category `json:"categories"`
}

type listExpenseGroupsResponse struct {
	ExpenseGroups []expenseGroupResponse `json:"expense_groups"`
}

type createExpenseGroupRequest struct {
	Name string `json:"name"`
}

type updateExpenseGroupRequest struct {
	Name string `json:"name"`
}

type createExpenseCategoryRequest struct {
	GroupID uuid.UUID `json:"group_id"`
	Name    string    `json:"name"`
}

type updateExpenseCategoryRequest struct {
	GroupID uuid.UUID `json:"group_id"`
	Name    string    `json:"name"`
}

type setCategoryDashboardRequest struct {
	Include *bool `json:"include"`
}

func writeExpenseCategoryError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, expensecategory.ErrNameRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, expensecategory.ErrGroupNotFound),
		errors.Is(err, expensecategory.ErrCategoryNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, expensecategory.ErrGroupNameTaken),
		errors.Is(err, expensecategory.ErrCategoryNameTaken),
		errors.Is(err, expensecategory.ErrGroupInactive),
		errors.Is(err, expensecategory.ErrGroupNotEmpty):
		writeError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (s *Server) handleListExpenseGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groups, err := s.expenseCategories.ListGroups(r.Context())
		if err != nil {
			log.Printf("list expense groups: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		categories, err := s.expenseCategories.ListCategories(r.Context())
		if err != nil {
			log.Printf("list expense categories: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		byGroup := make(map[uuid.UUID][]expensecategory.Category, len(groups))
		for _, c := range categories {
			byGroup[c.GroupID] = append(byGroup[c.GroupID], c)
		}

		out := make([]expenseGroupResponse, 0, len(groups))
		for _, g := range groups {
			cats := byGroup[g.ID]
			if cats == nil {
				cats = []expensecategory.Category{}
			}
			out = append(out, expenseGroupResponse{
				Group:      g,
				Categories: cats,
			})
		}

		writeJSON(w, http.StatusOK, listExpenseGroupsResponse{ExpenseGroups: out})
	}
}

func (s *Server) handleCreateExpenseGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createExpenseGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		created, err := s.expenseCategories.CreateGroup(r.Context(), req.Name)
		if err != nil {
			writeExpenseCategoryError(w, err, "create expense group")
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateExpenseGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req updateExpenseGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		updated, err := s.expenseCategories.UpdateGroup(r.Context(), id, req.Name)
		if err != nil {
			writeExpenseCategoryError(w, err, "update expense group")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeactivateExpenseGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.expenseCategories.SetGroupActive(r.Context(), id, false); err != nil {
			writeExpenseCategoryError(w, err, "deactivate expense group")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleActivateExpenseGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.expenseCategories.SetGroupActive(r.Context(), id, true); err != nil {
			writeExpenseCategoryError(w, err, "activate expense group")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleCreateExpenseCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createExpenseCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.GroupID == uuid.Nil {
			writeError(w, http.StatusBadRequest, "group_id is required")
			return
		}

		created, err := s.expenseCategories.CreateCategory(r.Context(), req.GroupID, req.Name)
		if err != nil {
			writeExpenseCategoryError(w, err, "create expense category")
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateExpenseCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req updateExpenseCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.GroupID == uuid.Nil {
			writeError(w, http.StatusBadRequest, "group_id is required")
			return
		}

		updated, err := s.expenseCategories.UpdateCategory(r.Context(), id, req.GroupID, req.Name)
		if err != nil {
			writeExpenseCategoryError(w, err, "update expense category")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeactivateExpenseCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.expenseCategories.SetCategoryActive(r.Context(), id, false); err != nil {
			writeExpenseCategoryError(w, err, "deactivate expense category")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleActivateExpenseCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.expenseCategories.SetCategoryActive(r.Context(), id, true); err != nil {
			writeExpenseCategoryError(w, err, "activate expense category")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleSetExpenseCategoryDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req setCategoryDashboardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Include == nil {
			writeError(w, http.StatusBadRequest, "include is required")
			return
		}

		if err := s.expenseCategories.SetCategoryInDashboard(r.Context(), id, *req.Include); err != nil {
			writeExpenseCategoryError(w, err, "set expense category dashboard")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
