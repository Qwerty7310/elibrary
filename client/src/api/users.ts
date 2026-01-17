import type {User} from "../types/library"
import {requestJson} from "./http"

export function getUserByID(id: string) {
    return requestJson<User>(`/admin/users/${encodeURIComponent(id)}`)
}

export function updateUser(id: string, payload: Partial<User> & {password?: string}) {
    return requestJson<void>(`/admin/users/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(payload),
    })
}

