import type {User} from "../types/library"
import {requestJson} from "./http"

export function getUserByID(id: string) {
    return requestJson<User>(`/admin/users/${encodeURIComponent(id)}`)
}

export function createUser(payload: {
    login: string
    first_name: string
    last_name?: string
    middle_name?: string
    email?: string
    password: string
    roles: User["roles"]
}) {
    return requestJson<User>("/admin/users", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

export function updateUser(id: string, payload: Partial<User> & {password?: string}) {
    return requestJson<void>(`/admin/users/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(payload),
    })
}
