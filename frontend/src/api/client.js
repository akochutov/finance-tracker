const BASE_URL = "http://localhost:8080"

async function request(path) {
    const response = await fetch(`${BASE_URL}${path}`);
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
}

export async function getCurrencies() {
    const data = await request("/api/currencies");
    return data.currencies
}

export async function getCompanies() {
    const data = await request("/api/companies");
    return data.companies;
}

export async function getIncomes() {
    const data = await request("/api/incomes");
    return data.incomes;
}