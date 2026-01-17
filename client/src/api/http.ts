const API_URL =
    window.location.hostname === "localhost"
        ? "http://localhost:8080"
        : "/elibrary/api"

export class ApiError extends Error {
    status?: number

    constructor(message: string, status?: number) {
        super(message)
        this.name = "ApiError"
        this.status = status
    }
}

export function getToken() {
    return localStorage.getItem("auth_token")
}

export function setToken(token: string | null) {
    if (token) {
        localStorage.setItem("auth_token", token)
    } else {
        localStorage.removeItem("auth_token")
    }
}

async function parseJsonSafe(res: Response) {
    const text = await res.text()
    if (!text) {
        return null
    }
    try {
        return JSON.parse(text)
    } catch {
        return null
    }
}

export async function requestJson<T>(
    path: string,
    options: RequestInit = {},
    withAuth = true
): Promise<T> {
    const headers = new Headers(options.headers)
    if (withAuth) {
        const token = getToken()
        if (token) {
            headers.set("Authorization", `Bearer ${token}`)
        }
    }
    if (options.body && !headers.has("Content-Type")) {
        headers.set("Content-Type", "application/json")
    }

    const res = await fetch(`${API_URL}${path}`, {
        ...options,
        headers,
    })

    if (!res.ok) {
        const data = await parseJsonSafe(res)
        const message =
            (data && typeof data === "object" && "error" in data
                ? String((data as {error: string}).error)
                : undefined) || `request failed: ${res.status}`
        throw new ApiError(message, res.status)
    }

    const data = await parseJsonSafe(res)
    return data as T
}

