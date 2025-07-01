package summarize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sirupsen/logrus"
)

var (
	colorGreen  renderer.Tint = renderer.Tint{BG: renderer.Colors{color.Bold}, FG: renderer.Colors{color.FgGreen}}
	colorYellow renderer.Tint = renderer.Tint{BG: renderer.Colors{color.Bold}, FG: renderer.Colors{color.FgHiYellow}}
	colorRed    renderer.Tint = renderer.Tint{BG: renderer.Colors{color.Bold}, FG: renderer.Colors{color.FgHiRed}}
	colorReset  renderer.Tint = renderer.Tint{BG: renderer.Colors{color.Reset}, FG: renderer.Colors{color.Reset}}
)

type resourceActions struct {
	action     string
	components []string
}

func printColored(str string) {
	logrus.Infof("\033[1;93m%s\033[0m", str)
}

func getResourceChanges(rawPlan map[string]interface{}, componentPath string) ([]string, []string, []string) {
	resourcesToChangeArray := []string{}
	resourcesActionArray := []string{}
	resourcesToChangeInComponentArray := []string{}

	//Extract resource_changes block
	resourceChanges := rawPlan["resource_changes"]
	if resourceChanges != nil {
		// sortutil.AscByField(resourceChanges, "address") => FIXME: does not work!
		// Loop over the block and get the name, type and change fields from tf json plan file
		for i := range resourceChanges.([]interface{}) {
			resource := resourceChanges.([]interface{})[i].(map[string]interface{})["address"].(string)
			//Action performed on Resource to change
			change := resourceChanges.([]interface{})[i].(map[string]interface{})["change"]
			changeAction := change.(map[string]interface{})["actions"]
			strChangeAction := fmt.Sprintf("%v", changeAction)
			//Assign to Array
			resourcesToChangeArray = append(resourcesToChangeArray, resource)
			resourcesActionArray = append(resourcesActionArray, strChangeAction)
			resourcesToChangeInComponentArray = append(resourcesToChangeInComponentArray, componentPath)

		}
		//Return the Array
		return resourcesToChangeArray, resourcesActionArray, resourcesToChangeInComponentArray
	}
	return nil, nil, nil
}

/*
====================================
Function to find terraform plan file
====================================
*/
func findPlanFiles(dir string) []string {
	tfPlansLoc := []string{}
	regExpr, e := regexp.Compile(`tfplan+\.(json)$`) // find plan files of the form "tfplan.json"
	if e != nil {
		logrus.Fatal("Error finding plan file using regex", e)
	}
	e = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && regExpr.MatchString(info.Name()) {
			tfPlansLoc = append(tfPlansLoc, path)
		}
		return nil
	})
	if e != nil {
		logrus.Fatal(e)
	}

	sort.Strings(tfPlansLoc)

	return tfPlansLoc
}

func actionExists(action string, mapValue []resourceActions) (result bool) {
	result = false
	for _, value := range mapValue {
		if value.action == action {
			result = true
			break
		}
	}
	return result
}

// Helper function to format counter values
func formatCounterAndColor(counter int) string {
	if counter == 0 {
		return "-"
	}
	return strconv.Itoa(counter)
}

func colorConfig(headers []renderer.Tint, footers []renderer.Tint, columns []renderer.Tint) renderer.ColorizedConfig {
	return renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG:      renderer.Colors{color.Bold},
			Columns: headers,
		},
		Footer: renderer.Tint{
			Columns: footers,
		},
		Column: renderer.Tint{
			Columns: columns,
		},
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenColumns: tw.On,
				BetweenRows:    tw.On,
			},
		},
		Symbols:   tw.NewSymbols(tw.StyleRounded),
		Border:    colorReset,
		Separator: colorReset,
	}

}

func tableConfig() tablewriter.Config {

	return tablewriter.Config{
		Header: tw.CellConfig{},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapNormal,
				// MergeMode: tw.MergeBoth,
			},
			ColMaxWidths: tw.CellWidth{Global: 32},
		},
	}
}

func summarizeDetailedPlan(planFiles []string, detailPlan string) error {
	resourceMappedArray := make(map[string][]resourceActions)
	for i := range planFiles {

		var rawPlan map[string]interface{}
		planName := strings.TrimSuffix(filepath.Base(planFiles[i]), ".tfplan.json")
		if planName != detailPlan {
			continue
		}
		componentPath := strings.ReplaceAll(planName, "__", "/")
		jsonFile, err := os.ReadFile(planFiles[i])
		if err != nil {
			logrus.Fatal("Error reading json plan: "+planFiles[i]+" =>", err)
			return err
		}
		err = json.Unmarshal(jsonFile, &rawPlan)
		if err != nil {
			logrus.Fatal("Error parsing json plan: "+planFiles[i]+" =>", err)
			return err
		}

		resourceChanges, resourceChangeActions, resourceChangedInComponents := getResourceChanges(rawPlan, componentPath)
		// Populate the consolitdated resource changes array
		for i := range resourceChanges {
			resource := resourceChanges[i]
			action := resourceChangeActions[i]
			component := resourceChangedInComponents[i]

			if actionMaps, ok := resourceMappedArray[resource]; ok {

				// check if the action is there in the array
				result := actionExists(action, actionMaps)
				if !result {
					resourceMappedArray[resource] = append(resourceMappedArray[resource], resourceActions{action, []string{component}})
				} else {
					for j := range actionMaps {
						if action == actionMaps[j].action {
							actionMaps[j].components = append(actionMaps[j].components, component)
							break
						}
					}
				}
			} else {
				resourceMappedArray[resource] = append(resourceMappedArray[resource], resourceActions{action, []string{component}})
			}
		}

		printColored(fmt.Sprintf("\nPROJECT CHANGES => %s", componentPath))

		tableString := &strings.Builder{}

		tableCfg := tableConfig()
		colorCfg := colorConfig(
			[]renderer.Tint{
				colorGreen, colorReset, colorReset, colorReset, colorReset, colorReset,
			},
			[]renderer.Tint{
				colorReset, colorReset, colorYellow, colorGreen, colorRed, colorRed,
			},
			[]renderer.Tint{
				colorReset, colorReset, colorReset, colorReset, colorReset, colorGreen,
			},
		)

		table := tablewriter.NewTable(tableString, tablewriter.WithRenderer(renderer.NewColorized(colorCfg)), tablewriter.WithConfig(tableCfg))
		table.Header([]string{"RESOURCE", "📖", "✏️", "🆕", "🗑", "🔄"})
		tableRowCounter := 0
		for i := range resourceChanges {
			if resourceChangeActions[i] == "[no-op]" {
				continue
			}
			readFlag, createFlag, deleteFlag, updateFlag, replacedFlag := "", "", "", "", ""

			switch resourceChangeActions[i] {
			case "[read]":
				readFlag = "-"
			case "[create]":
				createFlag = "-"
			case "[update]":
				updateFlag = "-"
			case "[delete]":
				deleteFlag = "-"
			case "[delete create]":
				replacedFlag = "-"
			case "[create delete]":
				replacedFlag = "-"
			}

			tableRow := make([]string, 0)
			tableRow = append(tableRow, resourceChanges[i])
			tableRow = append(tableRow, readFlag)
			tableRow = append(tableRow, createFlag)
			tableRow = append(tableRow, updateFlag)
			tableRow = append(tableRow, deleteFlag)
			tableRow = append(tableRow, replacedFlag)
			if err = table.Append(tableRow); err != nil {
				logrus.Fatal("Error appending row to table", err)
				return err
			}
			tableRowCounter++

		}

		if tableRowCounter == 0 {
			tableRow := []string{"N/A", "N/A", "N/A", "N/A", "N/A", "N/A"}
			if err := table.Append(tableRow); err != nil {
				logrus.Fatal("Error appending row to table", err)
				return err
			}
			table.Footer([]string{"-", "-", "-", "-", "-", "NO CHANGES"}) // Add Footer
		}
		if err := table.Render(); err != nil {
			logrus.Fatal("Error rendering the table", err)
			return err
		}
		logrus.Info(tableString.String())

	}
	return nil
}

func summarizeAllPlans(planFiles []string, envProjectRegex string) error {

	tableCfg := tableConfig()
	colorCfg := colorConfig(
		[]renderer.Tint{
			colorGreen, colorGreen, colorGreen, colorGreen, colorGreen, colorGreen, colorGreen, colorGreen,
		},
		[]renderer.Tint{
			colorRed, colorReset, colorReset, colorReset, colorReset, colorReset, colorReset, colorReset,
		},
		[]renderer.Tint{
			colorYellow, colorReset, colorReset, colorReset, colorYellow, colorGreen, colorRed, colorRed,
		},
	)

	tableString := &strings.Builder{}
	table := tablewriter.NewTable(tableString, tablewriter.WithRenderer(renderer.NewColorized(colorCfg)), tablewriter.WithConfig(tableCfg))
	table.Header([]string{"PLAN_FILE", "ENVIRONMENT", "PROJECT", "📖", "✏️", "🆕", "🗑", "🔄"})

	tableRowCounter := 0
	for i := range planFiles {

		planName := strings.TrimSuffix(filepath.Base(planFiles[i]), ".tfplan.json")
		componentPath := strings.ReplaceAll(planName, "__", "/")
		re := regexp.MustCompile(envProjectRegex)
		matches := re.FindStringSubmatch(componentPath)
		if len(matches) != 3 {
			err := fmt.Errorf("could not extract the environment and project name from the project name using the given regex: %q, %q", componentPath, envProjectRegex)
			logrus.Fatal(err)
			return err
		}
		environment := matches[1]
		project := matches[2]

		var rawPlan map[string]interface{}
		jsonFile, err := os.ReadFile(planFiles[i])
		if err != nil {
			logrus.Fatal("Error reading json plan: "+planFiles[i]+" =>", err)
			return err
		}
		err = json.Unmarshal(jsonFile, &rawPlan)
		if err != nil {
			logrus.Fatal("Error parsing json plan: "+planFiles[i]+" =>", err)
			return err
		}

		_, resourceActions, _ := getResourceChanges(rawPlan, componentPath)
		readCounter, addCounter, deleteCounter, modifiedCounter, replacedCounter := 0, 0, 0, 0, 0
		for _, action := range resourceActions {
			switch action {
			case "[read]":
				readCounter += 1
			case "[create]":
				addCounter += 1
			case "[update]":
				modifiedCounter += 1
			case "[delete]":
				deleteCounter += 1
			case "[delete create]":
				replacedCounter += 1
			case "[create delete]":
				replacedCounter += 1
			}
		}
		// Append formatted counters and determine colors dynamically
		formattedRead := formatCounterAndColor(readCounter)
		formattedAdd := formatCounterAndColor(addCounter)
		formattedDelete := formatCounterAndColor(deleteCounter)
		formattedModified := formatCounterAndColor(modifiedCounter)
		formattedReplaced := formatCounterAndColor(replacedCounter)

		tableRow := []string{planName, environment, project, formattedRead, formattedModified, formattedAdd, formattedDelete, formattedReplaced}
		if err = table.Append(tableRow); err != nil {
			logrus.Fatal("Error appending row to table", err)
			return err
		}
		tableRowCounter++
	}

	if tableRowCounter == 0 {
		tableRow := []string{"N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A"}
		if err := table.Append(tableRow); err != nil {
			logrus.Fatal("Error appending row to table", err)
			return err
		}
		table.Footer([]string{"NO PLAN FILE FOUND", "-", "-", "-", "-", "-", "-", "-"}) // Add Footer
	}
	if err := table.Render(); err != nil {
		logrus.Fatal("Error rendering the table", err)
		return err
	}
	logrus.Info(tableString.String())
	return nil
}

func Summarize(plansDir string, detailPlan string, envProjectRegex string) error {

	planFiles := findPlanFiles(plansDir)

	var err error
	if detailPlan == "" {
		err = summarizeAllPlans(planFiles, envProjectRegex)
	} else {
		err = summarizeDetailedPlan(planFiles, detailPlan)
	}
	return err
}
