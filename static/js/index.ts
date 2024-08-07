import htmx from "htmx.org"
//@ts-ignore
window.htmx = htmx;

import { For, Show, render } from 'solid-js/web';
import html from 'solid-js/html';
import { Accessor, Setter, createEffect, createMemo, createSignal } from 'solid-js';

const dbg = <T>(v: T): T => {
    console.log(v);
    return v;
}

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
        parentEl?.querySelector(`.option-${focusedOption()}`)?.scrollIntoView({ block: "center" });
    })

    const isEmpty = () => props.multiple ? props.selectedId.length === 0 : props.selectedId === "";


    return html`
        <div class="relative" ref=${(e: HTMLElement) => parentEl = e}>
            <input
                class=${() => `block w-full peer p-2 placeholder-black dark:placeholder-white focus:placeholder-gray-400 dark:focus:placeholder-slate-400 bg-transparent border dark:border-slate-600 rounded-md ${focused() ? 'rounded-b-none' : ''}`}
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
            if (filteredOptions()[0]) {
                setFocusedOption(filteredOptions()[0].value)
            }
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
    const onBlur = (e) => { }

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

const TextInputControl = (p: { label: string, value: string, type: string, onInput: (e: Event) => void }) => {
    return html`
        <label class="flex flex-col gap-2">
            <p>${p.label}</p>
            <input
                type=${() => p.type}
                class="block w-full py-2 px-4 rounded-md bg-transparent border border-slate-600"
                value=${() => p.value}
                onInput=${p.onInput}
            />
        </label>
    `
}

type StorageLocation = {
    id: number,
    bin: string,
}

type Data = {
    productOptions: Option[],
    storageLocationOptions: Option[],
    inventoryItems: InventoryItem[],
    storageLocationOptionsMap: Map<number, Option>,
    productOptionsMap: Map<number, Option>,
}

const InventoryDeductionInterface = (p: { productOptions: Option[], storageLocationOptions: Option[], inventoryItems: InventoryItem[], csrfToken: string }) => {
    const [
        inventoryToBeDeducted,
        setInventoryToBeDeducted,
    ] = createSignal<{ id: number, quantity: number }[]>([]);

    const selectedProductIds = createMemo(() => {
        return [...new Set(inventoryToBeDeducted().map(i => p.inventoryItems.find(ii => ii.Id === i.id)!.Product_Id))];
    })

    const productOptionsMap = createMemo(() => {
        const map: Map<number, Option> = new Map();
        for (const o of p.productOptions) {
            map.set(+o.value, o);
        }

        return map;
    });

    const storageLocationOptionsMap = createMemo(() => {
        const map: Map<number, Option> = new Map();
        for (const slo of p.storageLocationOptions) {
            map.set(+slo.value, slo);
        }

        return map;
    });

    const inventoryItemsMap = createMemo(() => {
        const map: Map<number, InventoryItem> = new Map();
        for (const ii of p.inventoryItems) {
            map.set(ii.Id, ii);
        }
        return map;
    })

    const unusedInventoryItems = createMemo(() => {
        return p.inventoryItems.flatMap(i => {
            const found = inventoryToBeDeducted().find(si => si.id === i.Id);
            if (!found) return [i];

            if (i.Quantity > found.quantity) {
                return [
                    {...i, Quantity: i.Quantity - found.quantity}
                ];
            }

            return [];
        })
    });


    const [selectorData, setSelectorData] = createSignal<RawSelectorData>({
        productId: "",
        quantityStr: "",
        inventoryItems: [],
    });

    return html`
		<div class="flex flex-grow">
            <${InventorySelector}
                data=${selectorData}
                setData=${setSelectorData}
                storageLocationOptionsMap=${storageLocationOptionsMap}
                productOptionsMap=${productOptionsMap}
                productOptions=${() => p.productOptions}
                storageLocationOptions=${() => p.storageLocationOptions}
                inventoryItems=${unusedInventoryItems}
                onSelect=${(data: SelectorData) => {
                    let remaining = data.quantity;
                    const newItems: { id: number, quantity: number }[] = [];
                    for (const itemId of data.inventoryItems) {
                        const item = p.inventoryItems.find(i => i.Id === itemId);
                        if (!item) throw new Error("Item should exist");
                        newItems.push({ id: itemId, quantity: item.Quantity > remaining ? remaining : item.Quantity });
                        remaining -= item.Quantity;
                        if (remaining <= 0) break;
                    }

                    const uncollapsedItems = [...inventoryToBeDeducted(), ...newItems];
                    const collapsedItems: typeof uncollapsedItems = [];

                    for (const item of uncollapsedItems) {
                        const dupeIndex = collapsedItems.findIndex(ci => ci.id === item.id);
                        if (dupeIndex !== -1) {
                            const dupe = collapsedItems[dupeIndex];
                            collapsedItems[dupeIndex] = {id: dupe.id, quantity: dupe.quantity + item.quantity};
                        } else {
                            collapsedItems.push(item);
                        }
                    }

                    setInventoryToBeDeducted(collapsedItems);
                }}
            />
            <div class="flex flex-col border-l border-white flex-1">
                <ul class="w-full">
                    <${For} each=${selectedProductIds}>
                        ${(productId: number) => {
                            const items = createMemo(() => inventoryToBeDeducted().filter(item => inventoryItemsMap().get(item.id)!.Product_Id === productId));
                            return html`
                                <li class="p-4">
                                    ${productOptionsMap().get(productId)!.text}
                                    <ul>
                                        <${For} each=${items}>
                                            ${(item: {id: number, quantity: number}) => html`
                                                <li class="flex">
                                                    <button
                                                        class="p-2 flex items-center"
                                                        onClick=${(e) => {
                                                            setInventoryToBeDeducted(inventoryToBeDeducted().filter(i => i !== item));

                                                            setSelectorData({
                                                                quantityStr: item.quantity.toString(),
                                                                productId: inventoryItemsMap().get(item.id)!.Product_Id.toString(),
                                                                inventoryItems: [item.id],
                                                            })
                                                        }}
                                                    >
                                                        <div class="icon-[heroicons-outline--arrow-left]">-</div>
                                                    </button>
                                                    <button
                                                        class="p-2 flex items-center"
                                                        onClick=${(e) => {
                                                            setInventoryToBeDeducted(inventoryToBeDeducted().filter(i => i !== item));
                                                        }}
                                                    >
                                                        <div class="icon-[heroicons-outline--trash]"></div>
                                                    </button>
                                                    <p class="p-2">${storageLocationOptionsMap().get(inventoryItemsMap().get(item.id)!.Storage_Location_Id)!.text}</p>
                                                    <p class="p-2">${item.quantity} of ${p.inventoryItems.find(s => s.Id == item.id)!.Quantity}</p>
                                                </li>
                                            `}
                                        <//>
                                    </ul>
                                <//>
                            `
                        }}
                    <//>
                <//>
                <div class="p-4 mt-auto">
                    <${Show} when=${() => inventoryToBeDeducted().length > 0}>
                        <form method="POST">
                            <input
                                type="hidden"
                                name="csrf_token"
                                value=${() => p.csrfToken}
                            />
                            <input
                                type="hidden"
                                name="json_deductions"
                                value=${() => JSON.stringify(inventoryToBeDeducted())}
                            />
                            <button 
                                type="submit"
                                class="bg-blue-400 rounded-md p-2 px-4 outline-none ring-slate-800 dark:ring-yellow-200 focus-visible:ring duration-200 disabled:bg-slate-800 shadow-md"
                            >
                                Deduct
                            <//>
                        <//>
                    <//>
                <//>
            <//>
		<//>
    `
}

type SelectorData = {
    productId: number,
    inventoryItems: number[],
    quantity: number,
}
type RawSelectorData = {
    productId: string,
    inventoryItems: number[],
    quantityStr: string,
}
const InventorySelector = (p: {
    productOptions: Option[],
    storageLocationOptions: Option[],
    inventoryItems: InventoryItem[],
    data: RawSelectorData,
    setData: Setter<RawSelectorData>,
    onSelect: (s: SelectorData) => void,
}) => {
    const quantity = () => {
        const v = parseInt(p.data.quantityStr);
        if (isNaN(v)) {
            return undefined
        }
        return v;
    }

    const together = createMemo(() => {
        return p.inventoryItems.map(i => ({
            ...i,
            Bin: p.storageLocationOptions.find(s => +s.value === i.Storage_Location_Id)!.text
        }));
    });

    const filtered = createMemo(() => {
        return together().filter(i => i.Product_Id === +p.data.productId)
    })

    const sum = () => p.data.inventoryItems.reduce((acc, curr) => p.inventoryItems.find(i => i.Id === curr)!.Quantity + acc, 0);
    const disabled = () => quantity() === undefined || !p.data.productId || sum() < quantity()!

    const updateToAppropriateInventoryItemSelection = () => {
        let remaining = quantity();
        if (remaining === undefined) {
            p.setData({...p.data, inventoryItems: []});
            return
        }
        const items: number[] = [];
        for (const filteredItem of filtered()) {
            const item = p.inventoryItems.find(i => i.Id === filteredItem.Id);
            if (!item) throw new Error("Item should exist");
            items.push(filteredItem.Id);
            remaining -= filteredItem.Quantity;
            if (remaining <= 0) break;
        }
        p.setData({...p.data, inventoryItems: items});
    }

    createEffect(() => {
        console.log(p.data)
    })

    return html`
        <form
            class="flex flex-col gap-4 p-4 w-72"
            on:submit=${(e) => {
                e.preventDefault();
                if (disabled()) return;

                const prevData = {
                    productId: +p.data.productId,
                    inventoryItems: p.data.inventoryItems,
                    quantity: quantity()!,
                }
                p.setData({
                    productId: "",
                    quantityStr: "",
                    inventoryItems: [],
                })
                p.onSelect(prevData);
            }}
        >
        <${TextInputControl}
            label="Quantity"
            value=${() => p.data.quantityStr}
            onInput=${(e) => {
                p.setData({...p.data, quantityStr: e.currentTarget.value});
                updateToAppropriateInventoryItemSelection();
            }}
            type="number"
        />
        <label class="flex flex-col gap-2">
            <p>Product</p>
            <${FuzzySelect}
                selectedId=${() => p.data.productId}
                setSelectedId=${(id) => {
                    p.setData({...p.data, productId: id});
                    updateToAppropriateInventoryItemSelection();
                }}
                options=${() => p.productOptions}
                multiple=${false}
                onBlur=${(e) => { }}
            />
        <//>

        <div class="flex-grow flex flex-wrap overflow-scroll p-4 rounded-md gap-4 items-start">
            <ul class="list-disc">
                <${For} each=${() => filtered()}>
                    ${(item: InventoryItem & { Bin: string }) => {
                        const selected = createMemo(() => {
                            return p.data.inventoryItems.includes(item.Id);
                        })

                        return html`
                            <li>
                                <label>
                                    <input
                                        class="mr-2"
                                        type="checkbox"
                                        name="storageLocations"
                                        checked=${() => dbg(p.data.inventoryItems.includes(item.Id))}
                                        value=${() => item.Id}
                                        onInput=${(e) => {
                                            const newSelected = !selected();
                                            const base = p.data.inventoryItems.filter(i => i != item.Id)
                                            p.setData({...p.data, inventoryItems: newSelected ? base.concat(item.Id) : base })
                                        }}
                                    />
                                    ${item.Bin} - ${item.Quantity}
                                <//>
                            <//>
                        `;
            }}
                <//>
            <//>
        <//>

        <div>
            <button
                disabled=${disabled}
                class="bg-blue-400 rounded-md p-2 px-4 outline-none ring-slate-800 dark:ring-yellow-200 focus-visible:ring duration-200 disabled:bg-slate-800 shadow-md"
            >
                ${() => disabled()
            ? `${sum()} of ${quantity() === undefined ? 0 : quantity()}`
            : `Add ${quantity()} (${sum()} selected) `
        }
            </button>
        </div>
    <//>
    `
}

type InventoryItem = {
    Id: number,
    Product_Id: number,
    Quantity: number,
    Batch_Number: number,
    Storage_Location_Id: number,
}

for (const el of document.querySelectorAll(".inventory-deduction-interface")) {
    const jsonData = JSON.parse(((el as HTMLElement).querySelector('.json-data') as HTMLElement).innerText.trim())
    const csrfToken = ((el as HTMLElement).querySelector('.csrf-token') as HTMLElement).innerText.trim();
    const productOptions = jsonData.ProductOptions.map(o => ({ value: o.Value, text: o.Text }));
    const storageLocationOptions = jsonData.StorageLocationOptions.map(o => ({ value: o.Value, text: o.Text }));
    const inventoryItems: InventoryItem[] = jsonData.InventoryItems;

    render(() => html`
        <${InventoryDeductionInterface}
            productOptions=${productOptions}
            storageLocationOptions=${storageLocationOptions}
            inventoryItems=${inventoryItems}
            csrfToken=${csrfToken}
        />
   `, el);
}
