package forms

import (
	"encoding/json"
	"fmt"

	"github.com/Jeffail/gabs"
	"github.com/bep/gr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/evt"
	"github.com/murdinc/awsmDashboard/helpers"
)

type LaunchConfigurationClassForm struct {
	*gr.This
}

// Implements the StateInitializer interface
func (l LaunchConfigurationClassForm) GetInitialState() gr.State {
	return gr.State{"querying": true, "error": "", "success": "", "step": 1}
}

// Implements the ComponentDidMount interface
func (l LaunchConfigurationClassForm) ComponentWillMount() {
	var class map[string]interface{}

	if l.Props().Interface("class") != nil {
		classJson := l.Props().Interface("class").([]byte)
		json.Unmarshal(classJson, &class)
	}

	l.SetState(class)
	l.SetState(gr.State{"querying": true})

	// Get our options for the form
	go func() {
		endpoint := "//localhost:8081/api/classes/" + l.Props().String("apiType") + "/options"
		resp, err := helpers.GetAPI(endpoint)
		if !l.IsMounted() {
			return
		}
		if err != nil {
			l.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
			return
		}

		l.SetState(gr.State{"classOptionsResp": resp, "querying": false})
	}()
}

func (l LaunchConfigurationClassForm) Render() gr.Component {

	state := l.State()
	props := l.Props()

	// Form placeholder
	response := el.Div()

	// Print any alerts
	helpers.ErrorElem(state.String("error")).Modify(response)
	helpers.SuccessElem(state.String("success")).Modify(response)

	if state.Int("step") == 1 {
		if state.Bool("querying") {
			gr.Text("Loading.......").Modify(response)
		} else {
			l.BuildClassForm(props.String("className"), state.Interface("classOptionsResp")).Modify(response)
		}

	} else if state.Int("step") == 2 {

		if state.Bool("querying") {
			gr.Text("Saving...").Modify(response)
		} else {

			buttons := el.Div(
				gr.CSS("btn-toolbar"),
			)

			// Back
			el.Button(
				evt.Click(l.backButton).PreventDefault(),
				gr.CSS("btn", "btn-secondary"),
				gr.Text("Back"),
			).Modify(buttons)

			// Done
			el.Button(
				evt.Click(l.doneButton).PreventDefault(),
				gr.CSS("btn", "btn-primary"),
				gr.Text("Done"),
			).Modify(buttons)

			buttons.Modify(response)
		}

	}

	return response
}

func (l LaunchConfigurationClassForm) BuildClassForm(className string, optionsResp interface{}) *gr.Element {

	state := l.State()
	props := l.Props()

	var classOptions map[string][]string
	jsonParsed, _ := gabs.ParseJSON(optionsResp.([]byte))
	classOptionsJson := jsonParsed.S("classOptions").Bytes()
	json.Unmarshal(classOptionsJson, &classOptions)

	classEdit := el.Div(
		el.Header3(gr.Text(className)),
		el.HorizontalRule(),
	)

	classEditForm := el.Form()

	textField("Version", "version", &state, l.storeValue).Modify(classEditForm) // number
	selectOne("Instance Class", "instanceClass", classOptions["metricName"], &state, l.storeValue).Modify(classEditForm)
	checkbox("Rotate", "rotate", &state, l.storeValue).Modify(classEditForm)
	if state.Bool("rotate") {
		textField("Retain", "retain", &state, l.storeValue).Modify(classEditForm) // number
	}
	selectMultiple("Regions", "regions", classOptions["regions"], &state, l.storeValue).Modify(classEditForm)

	classEditForm.Modify(classEdit)

	buttons := el.Div(
		gr.CSS("btn-toolbar"),
	)

	// Back
	el.Button(
		evt.Click(l.backButton).PreventDefault(),
		gr.CSS("btn", "btn-secondary"),
		gr.Text("Back"),
	).Modify(buttons)

	// Save
	el.Button(
		evt.Click(l.saveButton).PreventDefault(),
		gr.CSS("btn", "btn-primary"),
		gr.Text("Save"),
	).Modify(buttons)

	// Delete
	if props.Interface("hasDelete") != nil && props.Bool("hasDelete") {
		el.Button(
			evt.Click(l.deleteButton).PreventDefault(),
			gr.CSS("btn", "btn-danger", "pull-right"),
			gr.Text("Delete"),
		).Modify(buttons)
	}

	buttons.Modify(classEdit)

	return classEdit

}

func (l LaunchConfigurationClassForm) backButton(*gr.Event) {
	l.SetState(gr.State{"success": ""})
	l.Props().Call("backButton")
}

func (l LaunchConfigurationClassForm) doneButton(*gr.Event) {
	l.SetState(gr.State{"success": ""})
	l.Props().Call("hideAllModals")
}

func (l LaunchConfigurationClassForm) saveButton(*gr.Event) {
	l.SetState(gr.State{"querying": true, "step": 2})

	cfg := make(map[string]interface{})
	for key, _ := range l.State() {
		cfg[key] = l.State().Interface(key)
	}

	go func() {
		endpoint := "//localhost:8081/api/classes/" + l.Props().String("apiType") + "/name/" + l.Props().String("className")

		_, err := helpers.PutAPI(endpoint, cfg)
		if !l.IsMounted() {
			return
		}

		if err != nil {
			l.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint), "step": 1})
			return
		}

		l.SetState(gr.State{"querying": false, "success": "Class was saved", "error": ""})
	}()

}

func (l LaunchConfigurationClassForm) deleteButton(*gr.Event) {
	l.SetState(gr.State{"querying": true})

	go func() {
		endpoint := "//localhost:8081/api/classes/" + l.Props().String("apiType") + "/name/" + l.Props().String("className")

		_, err := helpers.DeleteAPI(endpoint)
		if !l.IsMounted() {
			return
		}

		if err != nil {
			l.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
			return
		}

		l.SetState(gr.State{"querying": false, "success": "Class was deleted", "error": "", "step": 2})
	}()
}

func (l LaunchConfigurationClassForm) storeValue(event *gr.Event) {
	id := event.Target().Get("id").String()
	inputType := event.Target().Get("type").String()

	switch inputType {

	case "checkbox":
		l.SetState(gr.State{id: event.Target().Get("checked").Bool()})

	case "select-one":
		l.SetState(gr.State{id: event.TargetValue()})

	case "select-multiple":
		var vals []string
		options := event.Target().Length()

		for i := 0; i < options; i++ {
			if event.Target().Index(i).Get("selected").Bool() && event.Target().Index(i).Get("id") != nil {
				vals = append(vals, event.Target().Index(i).Get("id").String())
			}
		}
		l.SetState(gr.State{id: vals})

	default: // text, at least
		l.SetState(gr.State{id: event.TargetValue()})

	}
}