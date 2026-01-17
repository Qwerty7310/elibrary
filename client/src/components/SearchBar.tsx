import {useState} from "react"

interface Props {
    onSearch: (search: string) => void
    maxLength?: number
    onLimitExceeded?: (limit: number) => void
    onReset?: () => void
    isLoading?: boolean
}

export function SearchBar({
    onSearch,
    maxLength,
    onLimitExceeded,
    onReset,
    isLoading = false,
}: Props) {
    const [value, setValue] = useState("")

    return (
        <form
            className="search-bar"
            onSubmit={(e) => {
                e.preventDefault()
                onSearch(value)
            }}
        >
            <input
                type="search"
                value={value}
                onChange={(e) => {
                    const next = e.target.value
                    if (maxLength && next.length > maxLength) {
                        setValue(next.slice(0, maxLength))
                        onLimitExceeded?.(maxLength)
                        return
                    }
                    setValue(next)
                }}
                placeholder="Название, автор или штрихкод"
            />
            <button type="submit" disabled={isLoading}>
                Найти
            </button>
            <button
                className="ghost-button"
                type="button"
                onClick={() => {
                    setValue("")
                    onReset?.()
                }}
                disabled={isLoading || value.length === 0}
            >
                Сброс
            </button>
        </form>
    )
}
