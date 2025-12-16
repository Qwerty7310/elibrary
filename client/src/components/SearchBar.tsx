import {useState} from "react";

interface Props {
    onSearch: (search: string) => void
}

export function SearchBar({onSearch}: Props) {
    const [value, setValue] = useState("");

    return (
        <form
            onSubmit={(e) => {
                e.preventDefault();
                onSearch(value);
            }}
        >
            <input
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="Search..."
            />
            <button type="submit">Search</button>
        </form>
    )
}