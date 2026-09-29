import { Routes, Route, NavLink } from "react-router-dom";
import CurrenciesPage from "./components/CurrenciesPage";
import CompaniesPage from "./components/CompaniesPage";
import CompanyPage from "./components/CompanyPage";
import IncomesPage from "./components/IncomesPage";
import SettingsPage from "./components/SettingsPage";
import ExchangeRatesPage from "./components/ExchangeRatesPage";
import DashboardPage from "./components/DashboardPage";
import ExpenseCategoriesPage from "./components/ExpenseCategoriesPage";
import ExpensesPage from "./components/ExpensesPage";
import ExpensesDashboardPage from "./components/ExpensesDashboardPage";

function App() {
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-title">
          <div className="sidebar-brand">Finance tracker</div>
          <div className="sidebar-sub">Household ledger</div>
        </div>

        <nav className="sidebar-nav">
          <div className="sidebar-section-label">Dashboards</div>
          <NavLink to="/" end className="nav-link">Incomes</NavLink>
          <NavLink to="/expenses-dashboard" className="nav-link">Expenses</NavLink>

          <div className="sidebar-section-label">Ledger</div>
          <NavLink to="/currencies" className="nav-link">Currencies</NavLink>
          <NavLink to="/companies" className="nav-link">Companies</NavLink>
          <NavLink to="/incomes" className="nav-link">Incomes</NavLink>
          <NavLink to="/expenses" className="nav-link">Expenses</NavLink>
          <NavLink to="/expense-categories" className="nav-link">Expense categories</NavLink>
          <NavLink to="/exchange-rates" className="nav-link">Exchange Rates</NavLink>

          <div className="sidebar-section-label">Configuration</div>
          <NavLink to="/settings" className="nav-link">Settings</NavLink>

          <div className="sidebar-section-label">Coming soon</div>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Meters</span>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Tariffs</span>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Reports</span>
        </nav>
      </aside>

      <main className="content">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/currencies" element={<CurrenciesPage />} />
          <Route path="/companies" element={<CompaniesPage />} />
          <Route path="/companies/:id" element={<CompanyPage />} />
          <Route path="/incomes" element={<IncomesPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/exchange-rates" element={<ExchangeRatesPage />} />
          <Route path="/expenses" element={<ExpensesPage />} />
          <Route path="/expenses-dashboard" element={<ExpensesDashboardPage />} />
          <Route path="/expense-categories" element={<ExpenseCategoriesPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;