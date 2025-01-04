package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

const SettingsLocation = "settings.json"

func LoadSettings(settings *abstract.Settings) error {
	data, err := os.ReadFile(SettingsLocation)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, settings)

	if err != nil {
		return err
	}

	return nil
}

func SaveSettings(settings *abstract.Settings) error {
	bs, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(SettingsLocation, bs, 0666)
}

const MarkersLocation = "markers.json"

func LoadMarkersFile(markers *abstract.Markers) error {
	data, err := os.ReadFile(MarkersLocation)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, markers)

	if err != nil {
		return err
	}

	return nil
}

func SaveMarkersFile(markers *abstract.Markers) error {
	bs, err := json.Marshal(markers)
	if err != nil {
		return err
	}

	return os.WriteFile(MarkersLocation, bs, 0666)
}

func PrereadLogsFile(path string) ([]abstract.Marker, error) {
	file, err := os.Open(path)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	if err != nil {
		return nil, err
	}

	var markers []abstract.Marker
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		event := parser.ParseLine(line)

		if event == nil {
			continue
		}

		if login, ok := event.Contents.(*abstract.Login); ok {
			markers = append(markers, abstract.Marker{
				Time: event.Time,
				Name: fmt.Sprintf("Logged in as %v", login.Name),
				User: login.Name,
			})
		}

		if markerLine, ok := event.Contents.(*abstract.MarkerLine); ok {
			markers = append(markers, abstract.Marker{
				Time: event.Time,
				Name: markerLine.Name,
				User: markerLine.User,
			})
		}
	}

	return markers, nil
}
