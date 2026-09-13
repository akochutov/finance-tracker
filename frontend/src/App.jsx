import { Routes, Route, NavLink } from "react-router-dom";
import CurrenciesPage from "./components/CurrenciesPage";
import CompaniesPage from "./components/CompaniesPage";
import CompanyPage from "./components/CompanyPage";
import IncomesPage from "./components/IncomesPage";
import SettingsPage from "./components/SettingsPage";
import ExchangeRatesPage from "./components/ExchangeRatesPage";

function HomePage() {
  return (
    <div>
      <div className="page-header">
        <h1>Dashboard</h1>
        <p>Coming soon.</p>
      </div>
    </div>
  );
}

function App() {
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-title">
          <div className="sidebar-brand">Finance tracker</div>
          <div className="sidebar-sub">Household ledger</div>
        </div>

        <nav className="sidebar-nav">
          <NavLink to="/" end className="nav-link">Home</NavLink>
          <NavLink to="/currencies" className="nav-link">Currencies</NavLink>
          <NavLink to="/companies" className="nav-link">Companies</NavLink>
          <NavLink to="/incomes" className="nav-link">Incomes</NavLink>
          <NavLink to="/settings" className="nav-link">Settings</NavLink>
          <NavLink to="/exchange-rates" className="nav-link">Exchange Rates</NavLink>

          <div className="sidebar-section-label">Coming soon</div>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Expenses</span>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Meters</span>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Tariffs</span>
          <span className="nav-link" style={{ color: "var(--color-neutral-500)", cursor: "default" }}>Reports</span>
        </nav>
      </aside>

      <main className="content">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/currencies" element={<CurrenciesPage />} />
          <Route path="/companies" element={<CompaniesPage />} />
          <Route path="/companies/:id" element={<CompanyPage />} />
          <Route path="/incomes" element={<IncomesPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/exchange-rates" element={<ExchangeRatesPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;