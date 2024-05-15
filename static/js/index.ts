type HProps = Record<string, string | number | Signal<string | number> | ((e: Event) => void)>
type HChildren = HTMLElement | string | Signal<HTMLElement | string | HTMLElement[]> | (HChildren)[]

const insertAfter = (el: HTMLElement, newNode: Node, after: Node | null) => {
    if (after?.nextSibling) {
        el.insertBefore(newNode, after.nextSibling);
    } else {
        el.appendChild(newNode);
    }
};

const processChildren = (el: HTMLElement, children: HChildren, after: Node | null): Node | null => {
    if (Array.isArray(children)) {
        let newAfter = after;
        for (const child of children) {
            newAfter = processChildren(el, child, newAfter)
        }
        return newAfter
    } else if (typeof children === "string") {
        const tn = document.createTextNode(children)
        insertAfter(el, tn, after)
        return tn
    } else if (isSignal(children)) {
        const beginMarker = document.createComment("Begin")
        const endMarker = document.createComment("End")
        insertAfter(el, beginMarker, after)
        insertAfter(el, endMarker, beginMarker)
        effect(() => {
            let startNode = beginMarker.nextSibling!
            while (startNode !== endMarker) {
                const n = startNode
                startNode = startNode.nextSibling!
                n.remove()
            }
            processChildren(el, children.track(), beginMarker)
        })
        return endMarker
    } else {
        insertAfter(el, children, after)
        return children
    }
}

const applyProps = (el: HTMLElement, props: HProps) => {
    for (const [k, v] of Object.entries(props)) {
        if (k.startsWith("on:")) {
            if (typeof v !== 'function') throw new Error("Invalid event handler")
            el.addEventListener(k.slice(3), v)
        } else if (typeof v === 'string' || typeof v === 'number') {
            if (k === "value") {
                (el as HTMLInputElement).value = v.toString();
            } else {
                el.setAttribute(k, v.toString())
            }
        } else if (isSignal(v)) {
            effect(() => {
                if (k === "value") {
                    (el as HTMLInputElement).value = v.track().toString()
                } else {
                    el.setAttribute(k, v.track().toString())
                }
            })
        } else {
            throw new Error("Invalid property")
        }
    }
}
const _h = (tagName: string, props: HProps, children: HChildren = []) => {
    const el = document.createElement(tagName)
    applyProps(el, props)
    processChildren(el, children, el.lastChild)
    return el
}

const h: Record<string, (props: HProps, children?: HChildren) => HTMLElement> = new Proxy({}, {
    get(_target, property: string) {
        return (props: HProps, children: HChildren = []) => _h(property, props, children)
    }
})

const observableSymbol = Symbol("observable")
const isSignal = (t: unknown): t is Signal<unknown> => {
    return (t as any)[observableSymbol];
}
type Signal<T> = {
    [observableSymbol]: true
    observe(fn: (v: T) => void, runInitial?: boolean): () => void;
    set(newV: T): void;
    track(): T;
}

type SignalContextItems = Set<Signal<any>>[]
class SignalContext {
    #ctx: SignalContextItems
    constructor() {
        this.#ctx = []
    }
    track(o: Signal<unknown>) {
        if (this.#ctx.length === 0) {
            return
        }

        this.#ctx[this.#ctx.length - 1].add(o)
    }
    newCtx() {
        this.#ctx.push(new Set())
    }
    pop() {
        return this.#ctx.pop()
    }
}
const observableContext = new SignalContext()

function assert(value: boolean): asserts value is true {
    if (!value) {
        throw new Error("Failed assertion")
    }
}

const observable = <T>(initialValue: T): Signal<T> => {
    let value = initialValue
    const observers: Set<(v: T) => void> = new Set();

    const o: Signal<T> = {
        [observableSymbol]: true,
        observe(fn, runInitial = true) {
            if (runInitial) {
                fn(value)
            }
            observers.add(fn)

            return () => {
                observers.delete(fn)
            }
        },
        set(newV) {
            const oldValue = value
            value = newV
            if (newV === oldValue) return;
            for (const observer of observers) {
                observer(value)
            }
        },
        track() {
            observableContext.track(o)
            return value
        }
    }
    return o
}

const derived = <T>(fn: () => T): Signal<T> => {
    observableContext.newCtx()

    const d = observable(fn())

    const ctxSet = observableContext.pop()
    assert(!!ctxSet)

    for (const ctxItem of ctxSet!) {
        ctxItem.observe(() => {
            d.set(fn())
        }, false)
    }

    return d
}

const effect = (fn: () => void) => {
    observableContext.newCtx()
    fn()
    const ctxSet = observableContext.pop()
    for (const ctxItem of ctxSet!) {
        ctxItem.observe(() => {
            fn()
        }, false)
    }
}

const FuzzySelectOptions = ({ $options, $focusedOption, $selectedId, onChange }: { $options: Signal<Option[]>, $selectedId: Signal<string>, onChange: (newSelectedId: string) => void, $focusedOption: Signal<string> }) => {
    return derived(() => $options.track().map(option => {
        const selected = derived(() => $selectedId.track() === option.value)
        return h.div({
            class: derived(() => `flex ${$focusedOption.track() === option.value ? 'bg-gray-300' : ''} ${selected.track() ? 'bg-gray-200' : ''}`),
            'on:mousedown': () => {
                onChange(option.value)
            },
        }, [
            h.div({ class: "p-2" }, [derived(() => selected.track() ? "✔" : "○︎")]),
            h.div({ class: "p-2" }, [option.text]),
        ])
    }))
}


type Option = { value: string, text: string }

const clamp = (min: number, v: number, max: number) => {
    if (v < min) return min
    if (v > max) return max
    return v
}

const FuzzySelect = ({ $options, $selectedId }: { $options: Signal<Option[]>, $selectedId: Signal<string> }) => {
    const focused = observable(false)
    const $inputText = observable("")
    const $selectedText = derived(() => $options.track().find(o => o.value === $selectedId.track())?.text || "")
    const $mouseDownOnItem = observable(false)
    const $focusedOption = observable($selectedId.track())
    const $filteredOptions = derived(() => $options.track().filter(o => o.text.toLowerCase().includes($inputText.track().toLowerCase())))

    return h.div(
        { class: "w-min relative" },
        [
            h.input(
                {
                    class: "peer p-4 placeholder-black focus:placeholder-gray-400",
                    value: $inputText,
                    placeholder: $selectedText,
                    'on:focus': (_e) => {
                        const e = _e as FocusEvent & { currentTarget: HTMLInputElement }
                        focused.set(true)
                        e.currentTarget.setSelectionRange(0, e.currentTarget.value.length)
                    },
                    'on:blur': (_e) => {
                        const e = _e as FocusEvent & { currentTarget: HTMLInputElement }
                        if (!$mouseDownOnItem.track()) {
                            $inputText.set("")
                            focused.set(false)
                        } else {
                            e.currentTarget.focus()
                        }
                        $mouseDownOnItem.set(false)
                    },
                    'on:input': (e) => {
                        $inputText.set((e.currentTarget as any).value)
                    },
                    'on:keydown': (_e) => {
                        const e = _e as KeyboardEvent;
                        if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                            const focusedIndex = $filteredOptions.track().findIndex(o => o.value === $focusedOption.track())
                            const newIndex = clamp(0, e.key === "ArrowUp" ? focusedIndex - 1 : focusedIndex + 1, $filteredOptions.track().length - 1)
                            $focusedOption.set($filteredOptions.track()[newIndex].value)
                            e.preventDefault()
                        } else if (e.key === "Enter") {
                            $selectedId.set($focusedOption.track())
                            e.preventDefault()
                        }
                    },
                }
            ),
            h.div(
                {
                    class: derived(() => `${focused.track() ? '' : 'hidden'} absolute top-full z-30 duration-500 flex flex-col bg-white shadow-md w-full rounded-b-md`)
                },
                FuzzySelectOptions({
                    $options: $filteredOptions,
                    $selectedId,
                    $focusedOption,
                    onChange: (newSelectedId) => {
                        $selectedId.set(newSelectedId)
                        $inputText.set("")
                        $mouseDownOnItem.set(true)
                    }
                })
            ),
        ]
    )
}

for (const el of document.querySelectorAll(".fuzzy-select")) {
    const select = el.querySelector("select")!
    select.classList.toggle("hidden")

    console.log(select.value)
    const $selectedId = observable(select.value);
    const $options = observable(Array.from(el.querySelectorAll("option")).map(op => ({
        text: op.innerText,
        value: op.value,
    })))

    let initial = true;
    effect(() => {
        console.log("RUnning this")
        select.value = $selectedId.track()
        if (!initial) {
            select.dispatchEvent(new Event('input', {
                bubbles: true,
                cancelable: true,
            }))
        } else {
            initial = false
        }
    })

    el
        .querySelector('.js-mount')!
        .appendChild(FuzzySelect({ $options, $selectedId }))
}
