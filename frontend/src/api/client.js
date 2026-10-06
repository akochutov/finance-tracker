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

export async function closeBankRequisite(companyId, requisiteId, validTo) {
    return request(`/api/companies/${companyId}/bank-requisites/${requisiteId}/close`, {
        method: "POST",
        body: JSON.stringify({ valid_to: validTo }),
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

export async function closeCryptoRequisite(companyId, requisiteId, validTo) {
    return request(`/api/companies/${companyId}/crypto-requisites/${requisiteId}/close`, {
        method: "POST",
        body: JSON.stringify({ valid_to: validTo }),
    });
}

// --- Incomes ---

export async function getIncomes() {
    const data = await request("/api/incomes");
    return data.incomes;
}

export async function createIncome(income) {
    return request("/api/incomes", {
        method: "POST",
        body: JSON.stringify(income),
    });
}

// --- Settings ---

export async function getSettings() {
    return request("/api/settings");
}

export async function updateSettings(baseCurrency) {
    return request("/api/settings", {
        method: "PUT",
        body: JSON.stringify({ base_currency: baseCurrency }),
    });
}

export async function updateExpenseBaseCurrency(code) {
    return request("/api/settings", {
        method: "PUT",
        body: JSON.stringify({ expense_base_currency: code }),
    });
}

// --- Exchange Rates ---

export async function getExchangeRates(currency = "") {
    const query = currency ? `?currency=${encodeURIComponent(currency)}` : "";
    const data = await request(`/api/exchange-rates${query}`);
    return data.exchange_rates;
}

export async function createExchangeRate({ currency, rateAt, rate }) {
    return request("/api/exchange-rates", {
        method: "POST",
        body: JSON.stringify({
            currency,
            rate_at: rateAt,
            rate,
        }),
    });
}

// --- Rate sources (provider config) ---

export async function getRateSources() {
    const data = await request("/api/rate-sources");
    return data.rate_sources;
}

export async function getRateProviders() {
    const data = await request("/api/rate-providers");
    return data.providers;
}

export async function updateRateSource(kind, fields) {
    return request(`/api/rate-sources/${kind}`, {
        method: "PUT",
        body: JSON.stringify(fields),
    });
}

export async function fetchRates(kind) {
    return request(`/api/rate-sources/${kind}/fetch`, {
        method: "POST",
    });
}

export async function startBackfill(kind) {
    return request(`/api/rate-sources/${kind}/backfill`, {
        method: "POST",
    });
}

export async function getBackfillStatus(kind) {
    return request(`/api/rate-sources/${kind}/backfill`);
}

// --- Dashboard ---

export async function getIncomeDashboard() {
    return request("/api/dashboard/incomes");
}

// --- Expenses dashboard ---
 
export async function getExpenseDashboard(from, to) {
    const params = new URLSearchParams({ from, to });
    return request(`/api/dashboard/expenses?${params}`);
}

// --- Expense groups & categories ---

export async function getExpenseGroups() {
    const data = await request("/api/expense-groups");
    return data.expense_groups;
}

export async function createExpenseGroup(name) {
    return request("/api/expense-groups", {
        method: "POST",
        body: JSON.stringify({ name }),
    });
}

export async function updateExpenseGroup(id, name) {
    return request(`/api/expense-groups/${id}`, {
        method: "PUT",
        body: JSON.stringify({ name }),
    });
}

export async function deactivateExpenseGroup(id) {
    return request(`/api/expense-groups/${id}`, {
        method: "DELETE",
    });
}

export async function activateExpenseGroup(id) {
    return request(`/api/expense-groups/${id}/activate`, {
        method: "POST",
    });
}

export async function createExpenseCategory(groupId, name) {
    return request("/api/expense-categories", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, name }),
    });
}

export async function updateExpenseCategory(id, groupId, name) {
    return request(`/api/expense-categories/${id}`, {
        method: "PUT",
        body: JSON.stringify({ group_id: groupId, name }),
    });
}

export async function deactivateExpenseCategory(id) {
    return request(`/api/expense-categories/${id}`, {
        method: "DELETE",
    });
}

export async function activateExpenseCategory(id) {
    return request(`/api/expense-categories/${id}/activate`, {
        method: "POST",
    });
}

export async function setExpenseCategoryDashboard(id, include) {
    return request(`/api/expense-categories/${id}/dashboard`, {
        method: "PUT",
        body: JSON.stringify({ include }),
    });
}

// --- Expenses ---
 
export async function getExpenses(from, to) {
    const params = new URLSearchParams();
    if (from) params.set("from", from);
    if (to) params.set("to", to);
    const query = params.toString() ? `?${params}` : "";
    const data = await request(`/api/expenses${query}`);
    return data.expenses;
}
 
export async function createExpense(expense) {
    return request("/api/expenses", {
        method: "POST",
        body: JSON.stringify(expense),
    });
}
 
export async function updateExpense(id, expense) {
    return request(`/api/expenses/${id}`, {
        method: "PUT",
        body: JSON.stringify(expense),
    });
}
 
export async function deleteExpense(id) {
    return request(`/api/expenses/${id}`, {
        method: "DELETE",
    });
}

export async function getExpenseSuggestions() {
    const data = await request("/api/expenses/suggestions");
    return data.suggestions;
}

// --- Utilities: service types, addresses, accounts, meters, readings ---
 
export async function getServiceTypes() {
    const data = await request("/api/service-types");
    return data.service_types;
}
 
export async function getAddresses() {
    const data = await request("/api/addresses");
    return data.addresses;
}
 
export async function createAddress(address) {
    return request("/api/addresses", {
        method: "POST",
        body: JSON.stringify({ address }),
    });
}
 
export async function updateAddress(id, address) {
    return request(`/api/addresses/${id}`, {
        method: "PUT",
        body: JSON.stringify({ address }),
    });
}
 
export async function deactivateAddress(id) {
    return request(`/api/addresses/${id}`, { method: "DELETE" });
}
 
export async function activateAddress(id) {
    return request(`/api/addresses/${id}/activate`, { method: "POST" });
}
 
export async function getUtilityAccounts() {
    const data = await request("/api/utility-accounts");
    return data.accounts;
}

// account: { address_id, service, number, zones: "single" | "day_night" }
export async function createUtilityAccount(account) {
    return request("/api/utility-accounts", {
        method: "POST",
        body: JSON.stringify(account),
    });
}

export async function updateUtilityAccount(id, number) {
    return request(`/api/utility-accounts/${id}`, {
        method: "PUT",
        body: JSON.stringify({ number }),
    });
}

export async function deactivateUtilityAccount(id) {
    return request(`/api/utility-accounts/${id}`, { method: "DELETE" });
}

export async function activateUtilityAccount(id) {
    return request(`/api/utility-accounts/${id}/activate`, { method: "POST" });
}

// The provider's split of billing months by zone, for day/night accounts.
export async function getZoneUsage(accountId) {
    const data = await request(`/api/utility-accounts/${accountId}/zone-usage`);
    return data.zone_usage;
}

// month: "YYYY-MM"; values: { day: "…", night: "…" }. Overwrites the month.
export async function setZoneUsage(accountId, month, values) {
    return request(`/api/utility-accounts/${accountId}/zone-usage`, {
        method: "PUT",
        body: JSON.stringify({ month, values }),
    });
}

export async function deleteZoneUsage(accountId, month) {
    return request(`/api/utility-accounts/${accountId}/zone-usage/${month}`, { method: "DELETE" });
}

export async function getMeters() {
    const data = await request("/api/meters");
    return data.meters;
}

// meter: { account_id, serial, installed_on, removed_on, initial_on, initial_value }
export async function createMeter(meter) {
    return request("/api/meters", {
        method: "POST",
        body: JSON.stringify(meter),
    });
}

// fields: { serial, installed_on, removed_on }
export async function updateMeter(id, fields) {
    return request(`/api/meters/${id}`, {
        method: "PUT",
        body: JSON.stringify(fields),
    });
}

// Hard delete: the meter goes with all its readings.
export async function deleteMeter(id) {
    return request(`/api/meters/${id}`, { method: "DELETE" });
}

// --- Readings ---

// The monthly round: { taken_on, readings: [{ meter_id, value }] }
export async function createReadings(round) {
    return request("/api/readings", {
        method: "POST",
        body: JSON.stringify(round),
    });
}

// The newest reading of every meter.
export async function getLatestReadings() {
    const data = await request("/api/readings/latest");
    return data.readings;
}

export async function getReadings(meterId) {
    const data = await request(`/api/meters/${meterId}/readings`);
    return data.readings;
}

// body: { taken_on, value }
export async function updateReading(meterId, readingId, body) {
    return request(`/api/meters/${meterId}/readings/${readingId}`, {
        method: "PUT",
        body: JSON.stringify(body),
    });
}

export async function deleteReading(meterId, readingId) {
    return request(`/api/meters/${meterId}/readings/${readingId}`, { method: "DELETE" });
}

// --- Tariffs ---

export async function getTariffs() {
    const data = await request("/api/tariffs");
    return data.tariffs;
}

export async function createTariff(tariff) {
    return request("/api/tariffs", {
        method: "POST",
        body: JSON.stringify(tariff),
    });
}

export async function updateTariff(id, fields) {
    return request(`/api/tariffs/${id}`, {
        method: "PUT",
        body: JSON.stringify(fields),
    });
}

export async function deleteTariff(id) {
    return request(`/api/tariffs/${id}`, { method: "DELETE" });
}