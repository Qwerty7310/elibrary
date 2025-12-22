import {useState} from "react"

interface Props {
    onSearch: (search: string) => void
    isLoading?: boolean
}

export function SearchBar({onSearch, isLoading = false}: Props) {
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
                onChange={(e) => setValue(e.target.value)}
                placeholder="Название, автор или штрихкод"
            />
            <button type="submit" disabled={isLoading}>
                Найти
            </button>
        </form>
    )
}
