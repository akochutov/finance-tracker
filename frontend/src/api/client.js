const BASE_URL = "http://localhost:8080"

async function request(path, options = {}) {
    const response = await fetch(`${BASE_URL}${path}`, {
        headers: { "Content-Type": "application/json" },
        ...options,
    });
    if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(errorBody.error || `HTTP ${response.status}`);
    }
    if (response.status === 204) {
        return null;
    }
    return response.json();
}

// --- Currencies ---

export async function getCurrencies() {
    const data = await request("/api/currencies");
    return data.currencies
}

export async function createCurrency(currency) {
    return request("/api/currencies", {
        method: "POST",
        body: JSON.stringify(currency),
    });
}

export async function updateCurrency(code, fields) {
    return request(`/api/currencies/${code}`, {
        method: "PUT",
        body: JSON.stringify(fields),
    });
}

export async function deactivateCurrency(code) {
    return request(`/api/currencies/${code}`, {
        method: "DELETE",
    });
}

// --- Companies ---

export async function getCompanies() {
    const data = await request("/api/companies");
    return data.companies;
}

export async function getCompany(id) {
    return request(`/api/companies/${id}`);
}

export async function createCompany(company) {
    return request("/api/companies", {
        method: "POST",
        body: JSON.stringify(company),
    });
}

export async function updateCompany(id, fields) {
    return request(`/api/companies/${id}`, {
        method: "PUT",
        body: JSON.stringify(fields),
    });
}

export async function deactivateCompany(id) {
    return request(`/api/companies/${id}`, {
        method: "DELETE",
    });
}

// --- Requisites ---

export async function getBankRequisites(companyId) {
    const data = await request(`/api/companies/${companyId}/bank-requisites`);
    return data.bank_requisites;
}

export async function createBankRequisite(companyId, requisite) {
    return request(`/api/companies/${companyId}/bank-requisites`, {
        method: "POST",
        body: JSON.stringify(requisite),
    });
}

export async function getCryptoRequisites(companyId) {
    const data = await request(`/api/companies/${companyId}/crypto-requisites`);
    return data.crypto_requisites;
}

export async function createCryptoRequisite(companyId, requisite) {
    return request(`/api/companies/${companyId}/crypto-requisites`, {
        method: "POST",
        body: JSON.stringify(requisite),
    });
}

// --- Incomes ---

export async function getIncomes() {
    const data = await request("/api/incomes");
    return data.incomes;
}