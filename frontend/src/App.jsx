import CurrenciesList from "./components/CurrenciesList";
import CompaniesList from "./components/CompaniesList";
import IncomesList from "./components/IncomesList";

function App() {
  return (
    <div>
      <h1>Finance tracker</h1>
      <CurrenciesList />
      <CompaniesList />
      <IncomesList />
    </div>
  );
}

export default App;