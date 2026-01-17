import type {LocationEntity} from "../types/library"
import {requestJson} from "./http"

export function createLocation(payload: {
    parent_id?: string
    type: string
    name: string
    address?: string
    description?: string
}) {
    return requestJson<LocationEntity>("/admin/locations", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}
