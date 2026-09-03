import CurrenciesPage from "./components/CurrenciesPage";
import CompaniesList from "./components/CompaniesList";
import IncomesList from "./components/IncomesList";

function App() {
  return (
    <div>
      <h1>Finance tracker</h1>
      <CurrenciesPage />
      <CompaniesList />
      <IncomesList />
    </div>
  );
}

export default App;