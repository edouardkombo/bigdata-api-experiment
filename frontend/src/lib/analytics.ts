// Analytics API client
export async function getAnomalies() {
    return fetch('/ml/anomalies').then(res => res.json());
}

export async function getForecast() {
    return fetch('/ml/forecast').then(res => res.json());
}
