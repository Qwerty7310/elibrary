import type {Permission, RoleWithPermissions} from "../types/library"
import {requestJson} from "./http"

export function getRoles() {
    return requestJson<RoleWithPermissions[]>("/admin/roles")
}

export function getPermissions() {
    return requestJson<Permission[]>("/admin/permissions")
}

export function createRole(payload: {
    code: string
    name: string
    permission_codes: string[]
}) {
    return requestJson<void>("/admin/roles", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}

