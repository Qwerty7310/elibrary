import {requestJson} from "./http"

export async function loginUser(login: string, password: string) {
    const data = await requestJson<{access_token: string}>(
        "/auth/login",
        {
            method: "POST",
            body: JSON.stringify({login, password}),
        },
        false
    )
    return data.access_token
}

