export interface Book {
    id: string;
    barcode: string;
    title: string;
    author: string;
    publisher?: string;
    year?: number;
    location?: string;
    extra?: Record<string, any>;
}