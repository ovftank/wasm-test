package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
)

const (
	baseButtonClass     = "py-4 px-4 text-white rounded-full font-medium text-2xl transform transition-all duration-200 hover:scale-105 active:scale-95 focus:outline-none select-none"
	numberButtonClass   = "bg-gray-800 hover:bg-gray-700"
	operatorButtonClass = "bg-amber-500 hover:bg-amber-400"
	functionButtonClass = "bg-gray-400 hover:bg-gray-300 text-black"
	inputClass          = "w-full px-4 py-2 bg-gray-800 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 mt-4"
	submitButtonClass   = "w-full py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 focus:outline-none focus:ring-2 focus:ring-purple-500 transition-all duration-200 mt-2 disabled:opacity-50 disabled:cursor-not-allowed"
	resultMessageId     = "result-message"
	shareButtonClass    = "w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200 mt-2 flex items-center justify-center gap-2"
	submitFormId        = "submit-form"
	flagInputId         = "flag-input"
	resultMessageClass  = "text-center text-sm"
	successMessageClass = "text-center text-sm text-green-500 font-bold mb-4"
	errorMessageClass   = "text-center text-sm text-red-500"
	submitFormClass     = "mt-6 space-y-3"
	shareContainerClass = "mt-4 animate-bounce"
	bodyDefaultClass    = "min-h-screen bg-black flex items-center justify-center p-6 transition-all duration-1000"
	bodySuccessClass    = "min-h-screen bg-gradient-to-r from-red-400 to-cyan-400 flex items-center justify-center p-6 transition-all duration-1000"
)

type ButtonStyle struct {
	text      string
	class     string
	colSpan   int
	operation string
}

var (
	currentNumber string
	operator      string
	firstNumber   float64
	newNumber     bool
	expression    string
	lastResult    string
	buttons       = []ButtonStyle{
		{"AC", functionButtonClass, 1, "clear"},
		{"⌫", functionButtonClass, 1, "backspace"},
		{"%", functionButtonClass, 1, "percent"},
		{"÷", operatorButtonClass, 1, "/"},
		{"7", numberButtonClass, 1, "7"},
		{"8", numberButtonClass, 1, "8"},
		{"9", numberButtonClass, 1, "9"},
		{"×", operatorButtonClass, 1, "*"},
		{"4", numberButtonClass, 1, "4"},
		{"5", numberButtonClass, 1, "5"},
		{"6", numberButtonClass, 1, "6"},
		{"-", operatorButtonClass, 1, "-"},
		{"1", numberButtonClass, 1, "1"},
		{"2", numberButtonClass, 1, "2"},
		{"3", numberButtonClass, 1, "3"},
		{"+", operatorButtonClass, 1, "+"},
		{"0", numberButtonClass, 2, "0"},
		{".", numberButtonClass, 1, "."},
		{"=", operatorButtonClass, 1, "calculate"},
	}
)
var buttonCallbacks map[string]js.Func
var keyboardMap = map[string]string{
	"0": "0", "1": "1", "2": "2", "3": "3", "4": "4",
	"5": "5", "6": "6", "7": "7", "8": "8", "9": "9",
	".": ".", "+": "+", "-": "-", "*": "*", "/": "/",
	"Enter": "calculate", "=": "calculate",
	"Escape": "clear", "Backspace": "backspace",
	"%": "percent",
	"x": "*", "X": "*",
}
var buttonElements = make(map[string]js.Value)

func setupCalculator() {
	document := js.Global().Get("document")
	keypad := document.Call("getElementById", "keypad")

	if keypad.IsUndefined() || keypad.IsNull() {
		return
	}

	for _, btn := range buttons {
		button := document.Call("createElement", "button")
		button.Set("innerHTML", btn.text)
		button.Set("className", fmt.Sprintf("%s %s", baseButtonClass, btn.class))
		if btn.colSpan > 1 {
			button.Get("style").Set("gridColumn", fmt.Sprintf("span %d", btn.colSpan))
		}
		buttonElements[btn.operation] = button
		operation := btn.operation
		callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			handleButtonClick(operation)
			return nil
		})
		if buttonCallbacks == nil {
			buttonCallbacks = make(map[string]js.Func)
		}
		buttonCallbacks[operation] = callback
		button.Call("addEventListener", "click", callback)
		keypad.Call("appendChild", button)
	}

	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(handleKeyDown))
	js.Global().Get("document").Call("addEventListener", "keyup", js.FuncOf(handleKeyUp))
	clearDisplay()

	resultMsg := document.Call("createElement", "div")
	resultMsg.Set("id", resultMessageId)
	resultMsg.Set("className", resultMessageClass)
	keypad.Get("parentNode").Call("appendChild", resultMsg)

	if js.Global().Get("localStorage").Call("getItem", "ctf_completed").String() == "true" {
		resultMsg.Set("textContent", "You have already completed this challenge!")
		resultMsg.Set("className", successMessageClass)
		addShareButton(document)
		return
	}

	submitForm := document.Call("createElement", "div")
	submitForm.Set("id", submitFormId)
	submitForm.Set("className", submitFormClass)

	input := document.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("id", flagInputId)
	input.Set("placeholder", "Enter your flag...")
	input.Set("className", inputClass)

	submitBtn := document.Call("createElement", "button")
	submitBtn.Set("innerHTML", "Submit Flag")
	submitBtn.Set("className", submitButtonClass)
	submitBtn.Set("disabled", true)
	submitBtn.Call("setAttribute", "data-hint", "Check the button's properties carefully...")

	submitCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handleSubmit()
		return nil
	})
	submitBtn.Call("addEventListener", "click", submitCallback)

	submitForm.Call("appendChild", input)
	submitForm.Call("appendChild", submitBtn)
	keypad.Get("parentNode").Call("appendChild", submitForm)
}

func handleButtonClick(operation string) {
	switch {
	case operation == "clear":
		clearDisplay()
	case operation == "backspace":
		backspace()
	case operation == "percent":
		calculatePercent()
	case operation == "calculate":
		calculate()
	case operation == "+" || operation == "-" || operation == "*" || operation == "/":
		setOperator(operation)
	default:
		appendNumber(operation)
	}
}

func handleKeyDown(this js.Value, args []js.Value) interface{} {
	event := args[0]
	key := event.Get("key").String()
	event.Call("preventDefault")
	if operation, ok := keyboardMap[key]; ok {
		if button, exists := buttonElements[operation]; exists {
			button.Get("classList").Call("add", "scale-95", "ring-2", "ring-purple-500", "ring-offset-2")
		}
		handleButtonClick(operation)
	}
	return nil
}

func handleKeyUp(this js.Value, args []js.Value) interface{} {
	event := args[0]
	key := event.Get("key").String()
	if operation, ok := keyboardMap[key]; ok {
		if button, exists := buttonElements[operation]; exists {
			button.Get("classList").Call("remove", "scale-95", "ring-2", "ring-purple-500", "ring-offset-2")
		}
	}
	return nil
}

func clearDisplay() {
	currentNumber = "0"
	operator = ""
	firstNumber = 0
	newNumber = true
	expression = ""
	lastResult = ""
	updateExpression()
}

func backspace() {
	if len(currentNumber) > 1 {
		currentNumber = currentNumber[:len(currentNumber)-1]
	} else {
		currentNumber = "0"
		newNumber = true
	}
	updateExpression()
}

func calculatePercent() {
	if num, err := strconv.ParseFloat(currentNumber, 64); err == nil {
		currentNumber = formatResult(num / 100)
		updateExpression()
	}
}

func appendNumber(digit string) {
	if newNumber {
		currentNumber = digit
		newNumber = false
	} else {
		if digit == "." && contains(currentNumber, ".") {
			return
		}
		currentNumber += digit
	}
	updateExpression()
}

func setOperator(op string) {
	if num, err := strconv.ParseFloat(currentNumber, 64); err == nil {
		firstNumber = num
		operator = op
		newNumber = true
		updateExpression()
	}
}

func calculate() {
	if operator == "" {
		return
	}

	secondNumber, err := strconv.ParseFloat(currentNumber, 64)
	if err != nil {
		return
	}

	var result float64
	document := js.Global().Get("document")
	flagInput := document.Call("getElementById", flagInputId)

	switch operator {
	case "+":
		result = firstNumber + secondNumber
		if firstNumber == 1337 && secondNumber == 1337 {
			if !flagInput.IsUndefined() && !flagInput.IsNull() {
				flagInput.Set("value", "Q1RGe000dGhf")
			}
			updateExpression()
			return
		}
	case "-":
		result = firstNumber - secondNumber
	case "*":
		result = firstNumber * secondNumber
		if result == 1787569 {
			if !flagInput.IsUndefined() && !flagInput.IsNull() {
				currentValue := flagInput.Get("value").String()
				if currentValue == "Q1RGe000dGhf" {
					flagInput.Set("value", currentValue+"MXNfRnVOfQ==")
				} else {
					flagInput.Set("value", "WTF???")
				}
			}
			updateExpression()
			return
		}
	case "/":
		if secondNumber != 0 {
			result = firstNumber / secondNumber
		} else {
			currentNumber = "Error"
			updateExpression()
			return
		}
	}

	lastResult = formatResult(result)
	currentNumber = lastResult
	operator = ""
	newNumber = true
	updateExpression()
}

func formatResult(num float64) string {
	str := strconv.FormatFloat(num, 'f', -1, 64)
	return str
}

func updateExpression() {
	document := js.Global().Get("document")
	display := document.Call("getElementById", "display")

	var displayText string
	if operator == "" {
		displayText = currentNumber
	} else {
		displayOp := operator
		switch operator {
		case "*":
			displayOp = "×"
		case "/":
			displayOp = "÷"
		}
		if newNumber {
			displayText = fmt.Sprintf("%.0f %s", firstNumber, displayOp)
		} else {
			displayText = fmt.Sprintf("%.0f %s %s", firstNumber, displayOp, currentNumber)
		}
	}
	if lastResult != "" {
		display.Set("innerHTML", lastResult)
		lastResult = ""
	} else {
		display.Set("innerHTML", displayText)
	}
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func handleSubmit() {
	document := js.Global().Get("document")
	input := document.Call("getElementById", flagInputId)
	resultMsg := document.Call("getElementById", resultMessageId)
	flagValue := input.Get("value").String()

	if flagValue == "CTF{M4th_1s_FuN}" {
		localStorage := js.Global().Get("localStorage")
		if localStorage.Call("getItem", "ctf_completed").String() != "true" {
			localStorage.Call("setItem", "ctf_completed", "true")
			resultMsg.Set("textContent", "🎉 Congratulations! You solved the challenge!")
			resultMsg.Set("className", successMessageClass)

			submitForm := document.Call("getElementById", submitFormId)
			submitForm.Get("style").Set("display", "none")

			addShareButton(document)

			document.Get("body").Set("className", bodySuccessClass)
			js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				document.Get("body").Set("className", bodyDefaultClass)
				return nil
			}), 3000)
		}
	} else {
		resultMsg.Set("textContent", "Incorrect flag. Keep trying!")
		resultMsg.Set("className", errorMessageClass)
		input.Set("className", inputClass)
	}
}

func addShareButton(document js.Value) {
	shareContainer := document.Call("createElement", "div")
	shareContainer.Set("className", shareContainerClass)

	shareBtn := document.Call("createElement", "button")
	shareBtn.Set("className", shareButtonClass)

	fbIcon := document.Call("createElement", "span")
	fbIcon.Set("innerHTML", `<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
		<path d="M22 12c0-5.523-4.477-10-10-10S2 6.477 2 12c0 4.991 3.657 9.128 8.438 9.878v-6.987h-2.54V12h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.46h-1.26c-1.243 0-1.63.771-1.63 1.562V12h2.773l-.443 2.89h-2.33v6.988C18.343 21.128 22 16.991 22 12z"/>
	</svg>`)

	shareText := document.Call("createElement", "span")
	shareText.Set("textContent", "Share on Facebook")

	shareBtn.Call("appendChild", fbIcon)
	shareBtn.Call("appendChild", shareText)
	shareContainer.Call("appendChild", shareBtn)

	shareBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		shareURL := "https://www.facebook.com/sharer/sharer.php?u=" + js.Global().Get("window").Get("location").Get("href").String()
		js.Global().Get("window").Call("open", shareURL, "_blank", "width=600,height=400")
		return nil
	}))

	document.Call("getElementById", resultMessageId).Get("parentNode").Call("appendChild", shareContainer)
}

func main() {
	done := make(chan struct{})
	setupCalculator()
	<-done
}
