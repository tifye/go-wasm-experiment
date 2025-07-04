//go:build js

package main

import (
	"tifye/go-wasm-test/render"
	"tifye/go-wasm-test/web"
)

func main() {
	counter := &Signal[int64]{
		value: 0,
		deps:  make([]Dependency[int64], 0),
	}
	items := &Signal[[]string]{
		value: make([]string, 0),
		deps:  make([]Dependency[[]string], 0),
	}
	text := &Signal[string]{
		value: "",
		deps:  make([]Dependency[string], 0),
	}

	renderer := web.NewDOMRenderer()

	Incrementer(renderer, IncrementerProps{
		counter: counter,
	})
	MyAmazingList(renderer, MyAmazingListProps{
		counter: counter,
		items:   items,
		text:    text,
	})

	select {}
}

type IncrementerProps struct {
	counter *Signal[int64]
}

func Incrementer(renderer render.Renderer, props IncrementerProps) {
	counter := props.counter

	// <button on:click={Increment}>increment</button>
	incBtn := renderer.NewComponent("button")
	incBtn.SetAttribute("innerText", "increment")
	incBtn.SetAttribute("on:click", func() {
		counter.Set(counter.Value() + 1)
	})
	renderer.Render(incBtn)

	// <button on:click={Decrement}>decrement</button>
	decBtn := renderer.NewComponent("button")
	decBtn.SetAttribute("innerText", "decrement")
	decBtn.SetAttribute("on:click", func() {
		counter.Set(counter.Value() - 1)
	})
	renderer.Render(decBtn)

	// <span>{counter}</span>
	lbl := renderer.NewComponent("span")
	lbl.SetAttribute("innerText", 0)
	counter.Effect(func() {
		lbl.SetAttribute("innerText", counter.Value())
	})
	renderer.Render(lbl)

	// <br />
	br := renderer.NewComponent("br")
	renderer.Render(br)
}

type MyAmazingListProps struct {
	counter *Signal[int64]
	items   *Signal[[]string]
	text    *Signal[string]
}

func MyAmazingList(renderer render.Renderer, props MyAmazingListProps) {
	counter := props.counter
	items := props.items
	text := props.text

	// <span>{text}</span>
	lbl := renderer.NewComponent("span")
	text.Effect(func() {
		lbl.SetAttribute("innerText", text.Value())
	})
	renderer.Render(lbl)

	// <input type="text" bind:value={text} value={counter} />
	input := renderer.NewComponent("input")
	input.SetAttribute("type", "text")
	text.Effect(func() {
		input.SetAttribute("value", text.Value())
	})
	input.SetAttribute("bind:value", func(val any) {
		str, ok := val.(string)
		if !ok {
			panic("not a string")
		}
		text.Set(str)
	})
	counter.Effect(func() {
		input.SetAttribute("value", counter.Value())
	})
	renderer.Render(input)

	// <button on:click={AddItem}>add item</button>
	addItemBtn := renderer.NewComponent("button")
	addItemBtn.SetAttribute("innerText", "add item")
	// todo: handle defer jsHandleAddItem.Release()
	addItemBtn.SetAttribute("on:click", func() {
		items.Set(append(items.Value(), input.(*web.WebComponent).Element().Get("value").String()))
	})
	renderer.Render(addItemBtn)

	/*
		<ul>
			@for _, item := range items {
				<li>{item}</li>
			}
		</ul>
	*/
	list := renderer.NewComponent("ul")
	items.Effect(func() {
		item := renderer.NewComponent("li")
		item.SetAttribute("innerText", items.Value()[len(items.Value())-1])
		renderer.Append(list, item)
	})
	renderer.Render(list)
}
