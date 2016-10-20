package components

import (
	"fmt"

	"github.com/bep/gr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/evt"
	"github.com/murdinc/awsmDashboard/helpers"
)

type AssetDropdownMenu struct {
	*gr.This
}

// Implements the StateInitializer interface.
func (d AssetDropdownMenu) GetInitialState() gr.State {
	return gr.State{"querying": false, "error": "", "classList": nil}
}

func (d AssetDropdownMenu) getClassList(event *gr.Event) {
	go func() {
		if apiType := d.Props().String("apiType"); apiType != "" {
			d.SetState(gr.State{"querying": true})
			endpoint := "//localhost:8081/api/classes/" + apiType
			resp, err := helpers.QueryAPI(endpoint)
			if !d.IsMounted() {
				return
			}
			if err != nil {
				d.SetState(gr.State{"querying": false, "error": fmt.Sprintf("Error while querying endpoint: %s", endpoint)})
				return
			}
			d.SetState(gr.State{"querying": false, "classList": resp})
		}
	}()
}

func (d AssetDropdownMenu) Render() gr.Component {

	state := d.State()
	props := d.Props()
	apiType := props.String("apiType")
	pageType := props.String("type")
	classList := state.Interface("classList")

	dropdown := el.Div(
		gr.CSS("btn-group", "dropdown"),
		el.Button(
			evt.Click(d.getClassList),
			gr.CSS("btn", "btn-primary", "btn-xs", "dropdown-toggle"),
			el.Italic(gr.CSS("fa", "fa-gear")),
			gr.Data("toggle", "dropdown"),
			gr.Aria("haspopup", "true"),
			gr.Aria("expanded", "false"),
		),
	)

	dropdownMenu := el.UnorderedList(
		gr.CSS("dropdown-menu"),
	)

	if state.Interface("classList") != nil {
		//el.ListItem(el.Anchor(gr.Data("toggle", "modal"), gr.Data("target", "#new-asset-modal"), gr.Text("New "+pageType))).Modify(dropdownMenu) // New Asset
		el.ListItem(el.Anchor(gr.Data("toggle", "modal"), gr.Data("target", "#new-class-modal"), gr.Text("New Class"))).Modify(dropdownMenu)   // New Class
		el.ListItem(el.Anchor(gr.Data("toggle", "modal"), gr.Data("target", "#edit-class-modal"), gr.Text("Edit Class"))).Modify(dropdownMenu) // Edit Classes

	} else if state.Bool("querying") {
		el.ListItem(gr.Text(" Loading...")).Modify(dropdownMenu)
	} else if errStr := state.String("error"); errStr != "" {
		el.ListItem(gr.Text(errStr)).Modify(dropdownMenu)
	}

	// New Asset
	// TODO
	/*
	   gr.New(&Modal{}).CreateElement(gr.Props{"id": "new-asset-modal", "title": "New " + pageType},
	   		gr.New(&NewAsset{}).CreateElement(gr.Props{"classList": classList, "apiType": apiType}),
	   	).Modify(dropdown)
	*/

	// Edit Class
	gr.New(&Modal{}).CreateElement(gr.Props{"id": "edit-class-modal", "title": "Edit " + pageType + " Classes"},
		gr.New(&EditClass{}).CreateElement(gr.Props{"classList": classList, "apiType": apiType}),
	).Modify(dropdown)

	// New Class
	gr.New(&Modal{}).CreateElement(gr.Props{"id": "new-class-modal", "title": "New " + pageType + " Class"},
		gr.New(&NewClass{}).CreateElement(gr.Props{"apiType": apiType}),
	).Modify(dropdown)

	dropdownMenu.Modify(dropdown)

	return dropdown
}
