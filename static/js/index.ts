import htmx from "htmx.org"
//@ts-ignore
window.htmx = htmx;

import { For, render } from 'solid-js/web';
import html from 'solid-js/html';
import { createEffect, createSignal } from 'solid-js';

const FuzzySelectOptions = (props: {
    options: Option[],
    multiple: true,
    selectedId: string[],
    onSelect: (id: string) => void,
    focusedOption: string,
} | {
    options: Option[],
    multiple: false,
    selectedId: string,
    onSelect: (id: string) => void,
    focusedOption: string,
}) => {
    return html`
        <${For} each=${() => props.options}>
            ${(option: Option) => {
            const selected = () => typeof props.selectedId === "string" ? (props.selectedId === option.value) : props.selectedId.includes(option.value)
            const className = () => `option-${option.value} flex ${props.focusedOption === option.value ? 'bg-gray-300 dark:bg-slate-800' : selected() ? 'bg-gray-200 dark:bg-slate-700' : ''}`
            return html`
                    <div
                        class=${className}
                        onMouseDown=${() => {
                            props.onSelect(option.value)
                        }}
                    >
                        <div class="p-2">
                            <div class=${() => selected() ? "icon-[heroicons-outline--check]" : "icon-[heroicons-outline--plus-circle]"} />
                        <//>
                        <div class="p-2">${option.text}<//>
                    <//>
                `
        }}
        <//>
    `;
}


type Option = { value: string, text: string }

const clamp = (min: number, v: number, max: number) => {
    if (v < min) return min
    if (v > max) return max
    return v
}

const FuzzySelect = (props: {
    options: Option[],
    selectedId: string[],
    multiple: true,
    setSelectedId: (newV: string[]) => void,
    onBlur: () => void,
} | {
    options: Option[],
    selectedId: string,
    multiple: false,
    setSelectedId: (newV: string) => void
    onBlur: () => void,
}) => {
    const [focused, setFocused] = createSignal(false);
    const [inputText, setInputText] = createSignal("");
    const selectedText = () => {
        if (!props.multiple) {
            return props.options.find(o => o.value === props.selectedId)?.text || "";
        }
        return props.selectedId.map(i => props.options.find(o => o.value === i)!.text).join(", ");
    }
    const [mouseDownOnItem, setMouseDownOnItem] = createSignal(false);
    const [focusedOption, setFocusedOption] = createSignal(typeof props.selectedId === "string" ? props.selectedId : props.selectedId[0]);
    const filteredOptions = () => {
        const baseFiltered = props.options.filter(o => o.text.toLowerCase().includes(inputText().toLowerCase()))
        if (!props.multiple) {
            return baseFiltered
        }
        return baseFiltered.concat(props.options.filter(o => props.selectedId.includes(o.value) && !baseFiltered.includes(o)))
    }

    const handleSelect = (newSelectedId: string) => {
        if (props.multiple) {
            props.setSelectedId(
                props.selectedId.includes(newSelectedId)
                    ? props.selectedId.filter(v => v != newSelectedId)
                    : props.selectedId.concat(newSelectedId)
            )
        } else {
            props.setSelectedId(newSelectedId)
        }
        if (!props.multiple) {
            setInputText("")
        }
    }

    const handleClear = () => {
        if (props.multiple) {
            props.setSelectedId([])
        } else {
            props.setSelectedId("")
        }
    }

    let parentEl: HTMLElement | null = null;
    createEffect(() => {
        parentEl?.querySelector(`.option-${focusedOption()}`)?.scrollIntoView({block: "center"});
    })

    const isEmpty = () => props.multiple ? props.selectedId.length === 0 : props.selectedId === "";
    

    return html`
        <div class="w-min relative" ref=${(e: HTMLElement) => parentEl = e}>
            <input
                class=${() => `peer p-2 placeholder-black dark:placeholder-white focus:placeholder-gray-400 dark:focus:placeholder-slate-400 bg-transparent border dark:border-slate-600 rounded-md ${focused() ? 'rounded-b-none' : ''}`}
                value=${inputText}
                placeholder=${selectedText}
                onFocus=${(e: FocusEvent & { currentTarget: HTMLInputElement }) => {
                    setFocused(true)
                    e.currentTarget.setSelectionRange(0, e.currentTarget.value.length)
                }}
                onBlur=${(e: FocusEvent & { currentTarget: HTMLInputElement }) => {
                    if (!mouseDownOnItem()) {
                        setInputText("")
                        setFocused(false)
                        props.onBlur()
                    } else {
                        e.currentTarget.focus()
                    }
                    setMouseDownOnItem(false)
                }}
                onInput=${(e: FocusEvent & { currentTarget: HTMLInputElement }) => {
                    setInputText(e.currentTarget.value)
                    setFocusedOption(filteredOptions()[0].value)
                }}
                onKeydown=${(e: KeyboardEvent) => {
                    if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                        const focusedIndex = filteredOptions().findIndex(o => o.value === focusedOption())
                        const newIndex = clamp(0, e.key === "ArrowUp" ? focusedIndex - 1 : focusedIndex + 1, filteredOptions().length - 1)
                        setFocusedOption(filteredOptions()[newIndex].value)
                        e.preventDefault()
                    } else if (e.key === "Enter") {
                        handleSelect(focusedOption())
                        e.preventDefault()
                    }
                }}
            />
            <div
                class=${() => `${focused() ? '' : 'hidden'} absolute top-full z-30 duration-500 flex flex-col dark:bg-slate-600 bg-white shadow-md dark:shadow-slate-400 w-full rounded-b-md max-h-[21.5rem] overflow-auto`}
            >
                <${FuzzySelectOptions}
                    multiple=${() => props.multiple}
                    options=${filteredOptions}
                    selectedId=${() => props.selectedId}
                    focusedOption=${focusedOption}
                    onSelect=${(newSelectedId: string) => {
                        handleSelect(newSelectedId)
                        setMouseDownOnItem(true)
                    }}
                />
            <//>
            <button
                type="button"
                class=${() => `${isEmpty() ? 'opacity-0 pointer-events-none' : ''} absolute right-2 top-1/2 transform -translate-y-1/2 bg-slate-800 p-1 duration-200`}
                tabindex=${() => isEmpty() ? -1 : 0}
                onClick=${handleClear}
            >
                <span class="icon-[heroicons-outline--x-mark]" />
            <//>
        <//>
    `
}

for (const el of document.querySelectorAll(".fuzzy-select")) {
    const select = el.querySelector("select")!;
    select.classList.toggle("hidden");
    const multiple = select.multiple;
    const onBlur = () => {}

    if (multiple) {
        const [selectedId, setSelectedId] = createSignal(Array.from(select.selectedOptions).map(o => o.value));
        const [options, _setOptions] = createSignal(Array.from(el.querySelectorAll("option")).map(op => ({
            text: op.innerText,
            value: op.value,
        })))

        let initial = true;
        createEffect(() => {
            for (const o of select.options) {
                o.selected = selectedId().includes(o.value);
            }
            if (!initial) {
                select.dispatchEvent(new Event('input', {
                    bubbles: true,
                    cancelable: true,
                }))
            } else {
                initial = false
            }
        })

        render(() => html`
            <${FuzzySelect}
                multiple=${true}
                options=${options}
                selectedId=${selectedId}
                setSelectedId=${setSelectedId}
                onBlur=${onBlur}
            />
       `, el.querySelector('.js-mount')!)
    } else {
        const [selectedId, setSelectedId] = createSignal(select.value);
        const [options, _setOptions] = createSignal(Array.from(el.querySelectorAll("option")).map(op => ({
            text: op.innerText,
            value: op.value,
        })))

        let initial = true;
        createEffect(() => {
            select.value = selectedId()
            if (!initial) {
                select.dispatchEvent(new Event('input', {
                    bubbles: true,
                    cancelable: true,
                }))
            } else {
                initial = false
            }
        })

        render(() => html`
            <${FuzzySelect}
                multiple=${false}
                options=${options}
                selectedId=${selectedId}
                setSelectedId=${setSelectedId}
                onBlur=${onBlur}
            />
       `, el.querySelector('.js-mount')!)
    }
}
