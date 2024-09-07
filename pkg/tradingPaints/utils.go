package tradingPaints

import (
	"encoding/xml"
	"fmt"
)

func parseData(data []byte) ([]Car, error) {
	var parsedData TPXML
	if err := xml.Unmarshal(data, &parsedData); err != nil {
		return nil, fmt.Errorf("fail parse XML: %v", err)
	}
	return parsedData.Cars.Car, nil
}
