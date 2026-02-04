import {requestJson} from "./http"

export type PrintTaskPayload = {
    str1: string
    str2: string
    barcode: string
}

export async function sendPrintTask(payload: PrintTaskPayload) {
    await requestJson<void>("/admin/print", {
        method: "POST",
        body: JSON.stringify(payload),
    })
}
