import {requestForm} from "./http"

export type ImageEntity = "author" | "book" | "publisher"

export async function uploadImage(
    entity: ImageEntity,
    id: string,
    file: File
): Promise<string> {
    const form = new FormData()
    form.append("image", file)
    const data = await requestForm<{url: string}>(`/admin/${entity}/${id}/image`, form)
    return data.url
}
