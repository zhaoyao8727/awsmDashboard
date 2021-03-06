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

type ImageClassForm struct {
	*gr.This
}

// Implements the StateInitializer interface
func (i ImageClassForm) GetInitialState() gr.State {
	return gr.State{"querying": true, "queryingOpts": true, "queryingInstances": true, "error": "", "success": "", "step": 1,
		"propagate": false,
		"rotate":    false,
		"instance":  "",
	}
}

// Implements the ComponentWillMount interface
func (i ImageClassForm) ComponentWillMount() {
	var class map[string]interface{}

	if i.Props().Interface("class") != nil {
		classJson := i.Props().Interface("class").([]byte)
		json.Unmarshal(classJson, &class)
	}

	i.SetState(class)
	i.SetState(gr.State{"queryingOpts": true, "queryingInstances": true})

	// Get our options for the form
	go func() {
		endpoint := "//localhost:8081/api/classes/" + i.Props().String("apiType") + "/options"
		resp, err := helpers.GetAPI(endpoint)
		if !i.IsMounted() {
			return
		}
		if err != nil {
			i.SetState(gr.State{"queryingOpts": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
			return
		}

		i.SetState(gr.State{"classOptionsResp": resp, "queryingOpts": false})
	}()

	// Get our existing instances for the form
	go func() {
		endpoint := "//localhost:8081/api/assets/instances-running"
		resp, err := helpers.GetAPI(endpoint)
		if !i.IsMounted() {
			return
		}
		if err != nil {
			i.SetState(gr.State{"queryingInstances": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
			return
		}

		i.SetState(gr.State{"instanceOptionsResp": resp, "queryingInstances": false})
	}()
}

func (i ImageClassForm) Render() gr.Component {

	state := i.State()
	props := i.Props()

	// Form placeholder
	response := el.Div()

	// Print any alerts
	helpers.ErrorElem(state.String("error")).Modify(response)
	helpers.SuccessElem(state.String("success")).Modify(response)

	if state.Int("step") == 1 {
		if state.Bool("queryingOpts") || state.Bool("queryingInstances") {
			gr.Text("Loading...").Modify(response)
		} else {
			i.BuildClassForm(props.String("className"), state.Interface("classOptionsResp"), state.Interface("instanceOptionsResp")).Modify(response)
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
				evt.Click(i.backButton).PreventDefault(),
				gr.CSS("btn", "btn-secondary"),
				gr.Text("Back"),
			).Modify(buttons)

			// Done
			el.Button(
				evt.Click(i.doneButton).PreventDefault(),
				gr.CSS("btn", "btn-primary"),
				gr.Text("Done"),
			).Modify(buttons)

			buttons.Modify(response)
		}

	}

	return response
}

func (i ImageClassForm) BuildClassForm(className string, optionsResp interface{}, instanceResp interface{}) *gr.Element {

	state := i.State()
	props := i.Props()

	var classOptions map[string][]string
	jsonParsed, _ := gabs.ParseJSON(optionsResp.([]byte))
	classOptionsJson := jsonParsed.S("classOptions").Bytes()
	json.Unmarshal(classOptionsJson, &classOptions)

	instanceJsonParsed, _ := gabs.ParseJSON(instanceResp.([]byte))
	instanceOptionsSlice, _ := instanceJsonParsed.S("assets").Children()

	var instances []string
	instancesMeta := make(map[string]string)
	for _, instanceOption := range instanceOptionsSlice {
		instance := instanceOption.S("instanceID").Data().(string)
		if instance != "" {
			instances = append(instances, instance)
			instancesMeta[instance] = instanceOption.S("name").Data().(string) + " " + instanceOption.S("availabilityZone").Data().(string)
		}
	}

	classEdit := el.Div(
		el.Header3(gr.Text(className)),
		el.HorizontalRule(),
	)

	classEditForm := el.Form(evt.KeyDown(DisableEnter))

	NumberField("Version", "version", state.Int("version"), i.storeValue).Modify(classEditForm)
	CreateableSelectMeta("Instance", "instance", instances, instancesMeta, state.Interface("instance"), i.storeSelect).Modify(classEditForm)
	Checkbox("Propagate", "propagate", state.Bool("propagate"), i.storeValue).Modify(classEditForm)
	if state.Bool("propagate") {
		SelectMultiple("Propagate Regions", "propagateRegions", classOptions["regions"], state.Interface("propagateRegions"), i.storeSelect).Modify(classEditForm)
	}
	Checkbox("Rotate", "rotate", state.Bool("rotate"), i.storeValue).Modify(classEditForm)
	if state.Bool("rotate") {
		NumberField("Retain", "retain", state.Int("retain"), i.storeValue).Modify(classEditForm)
	}

	classEditForm.Modify(classEdit)

	buttons := el.Div(
		gr.CSS("btn-toolbar"),
	)

	// Back
	el.Button(
		evt.Click(i.backButton).PreventDefault(),
		gr.CSS("btn", "btn-secondary"),
		gr.Text("Back"),
	).Modify(buttons)

	// Save
	el.Button(
		evt.Click(i.saveButton).PreventDefault(),
		gr.CSS("btn", "btn-primary"),
		gr.Text("Save"),
	).Modify(buttons)

	// Delete
	if props.Interface("hasDelete") != nil && props.Bool("hasDelete") {
		el.Button(
			evt.Click(i.deleteButton).PreventDefault(),
			gr.CSS("btn", "btn-danger", "pull-right"),
			gr.Text("Delete"),
		).Modify(buttons)
	}

	buttons.Modify(classEdit)

	return classEdit

}

func (i ImageClassForm) backButton(*gr.Event) {
	i.SetState(gr.State{"success": ""})
	i.Props().Call("backButton")
}

func (i ImageClassForm) doneButton(*gr.Event) {
	i.SetState(gr.State{"success": ""})
	i.Props().Call("hideAllModals")
}

func (i ImageClassForm) saveButton(*gr.Event) {
	i.SetState(gr.State{"querying": true, "step": 2})

	cfg := make(map[string]interface{})
	for key, _ := range i.State() {
		cfg[key] = i.State().Interface(key)
	}

	go func() {
		endpoint := "//localhost:8081/api/classes/" + i.Props().String("apiType") + "/name/" + i.Props().String("className")

		_, err := helpers.PutAPI(endpoint, cfg)
		if !i.IsMounted() {
			return
		}

		if err != nil {
			i.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint), "step": 1})
			return
		}

		i.SetState(gr.State{"querying": false, "success": "Class was saved", "error": ""})
	}()

}

func (i ImageClassForm) deleteButton(*gr.Event) {
	i.SetState(gr.State{"querying": true})

	go func() {
		endpoint := "//localhost:8081/api/classes/" + i.Props().String("apiType") + "/name/" + i.Props().String("className")

		_, err := helpers.DeleteAPI(endpoint)
		if !i.IsMounted() {
			return
		}

		if err != nil {
			i.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
			return
		}

		i.SetState(gr.State{"querying": false, "success": "Class was deleted", "error": "", "step": 2})
	}()
}

func (i ImageClassForm) storeValue(event *gr.Event) {
	key := event.Target().Get("name").String()
	inputType := event.Target().Get("type").String()

	switch inputType {

	case "checkbox":
		i.SetState(gr.State{key: event.Target().Get("checked").Bool()})

	case "number":
		i.SetState(gr.State{key: event.TargetValue().Int()})

	default: // text, at least
		i.SetState(gr.State{key: event.TargetValue()})

	}
}

func (i ImageClassForm) storeSelect(key string, val interface{}) {
	switch value := val.(type) {

	case map[string]interface{}:
		// single
		i.SetState(gr.State{key: value["value"]})

	case []interface{}:
		// multi
		var vals []string
		options := len(value)
		for i := 0; i < options; i++ {
			vals = append(vals, value[i].(map[string]interface{})["value"].(string))
		}
		i.SetState(gr.State{key: vals})

	default:
		i.SetState(gr.State{key: val})

	}
}
