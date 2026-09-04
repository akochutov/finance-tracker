import { Routes, Route, Link } from "react-router-dom";
import CurrenciesPage from "./components/CurrenciesPage";
import CompaniesPage from "./components/CompaniesPage";
import IncomesList from "./components/IncomesList";

function HomePage() {
  return <div>Dashboard - coming soon</div>;
}

function IncomesPage() {
  return (
    <div>
      <IncomesList />
    </div>
  );
}

function App() {
  return (
    <div>
      <h1>Finance tracker</h1>
      <nav>
        <Link to="/">Home</Link> |{" "}
        <Link to="/currencies">Currencies</Link> |{" "}
        <Link to="/companies">Companies</Link> |{" "}
        <Link to="/incomes">Incomes</Link>
      </nav>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/currencies" element={<CurrenciesPage />} />
        <Route path="/companies" element={<CompaniesPage />} />
        <Route path="/incomes" element={<IncomesPage />} />
      </Routes>
    </div>
  );
}

export default App;