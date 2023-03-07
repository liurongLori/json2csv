package json2csv

import (
	"encoding/csv"
	"io"
	"sort"

	"github.com/yukithm/json2csv/jsonpointer"
)

// KeyStyle represents the specific style of the key.
type KeyStyle uint

// Header style
const (
	// "/foo/bar/0/baz"
	JSONPointerStyle KeyStyle = iota

	// "foo/bar/0/baz"
	SlashStyle

	// "foo.bar.0.baz"
	DotNotationStyle

	// "foo.bar[0].baz"
	DotBracketStyle
)

// CSVWriter writes CSV data.
type CSVWriter struct {
	*csv.Writer
	HeaderStyle KeyStyle
	Transpose   bool
}

// NewCSVWriter returns new CSVWriter with given JSONPointerStyle and transpose.
func NewCSVWriter(w io.Writer, style KeyStyle, transpose bool) *CSVWriter {
	return &CSVWriter{
		csv.NewWriter(w),
		style,
		transpose,
	}
}

// WriterHeader only writes header.
func (w *CSVWriter) WriterHeader(csvHeader CSVHeader) error {
	header, err := w.FormatHeader(csvHeader)
	if err != nil {
		return err
	}

	if err := w.Write(header); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

// FormatHeader formats the given header with CSVWriter.HeaderStyle.
func (w *CSVWriter) FormatHeader(csvHeader CSVHeader) ([]string, error) {
	result := KeyValue{}
	for h := range csvHeader {
		result[h] = ""
	}
	results := []KeyValue{result}
	pts, err := allPointers(results)
	if err != nil {
		return nil, err
	}
	sort.Sort(pts)
	header := w.getHeader(pts)
	return header, nil
}

// WriteCSVByHeader writes CSV rows according the given header.
// For header columns of csvHeader that are missing in results, output an empty value.
// Fields of results that are absent in csvHeader are ignored.
func (w *CSVWriter) WriteCSVByHeader(results []KeyValue, csvHeader CSVHeader) error {
	result := KeyValue{}
	for h := range csvHeader {
		result[h] = ""
	}
	pts, err := allPointers([]KeyValue{result})
	if err != nil {
		return err
	}
	sort.Sort(pts)
	keys := pts.Strings()

	for _, result := range results {
		for h := range csvHeader {
			if _, exist := result[h]; !exist {
				result[h] = ""
			}
		}
		record := toRecord(result, keys)
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

// WriteCSV writes CSV data.
func (w *CSVWriter) WriteCSV(results []KeyValue) error {
	if w.Transpose {
		return w.writeTransposedCSV(results)
	}
	return w.writeCSV(results)
}

// WriteCSV writes CSV data.
func (w *CSVWriter) writeCSV(results []KeyValue) error {
	pts, err := allPointers(results)
	if err != nil {
		return err
	}
	sort.Sort(pts)
	keys := pts.Strings()
	header := w.getHeader(pts)

	if err := w.Write(header); err != nil {
		return err
	}

	for _, result := range results {
		record := toRecord(result, keys)
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

// WriteCSV writes CSV data which is transposed rows and columns.
func (w *CSVWriter) writeTransposedCSV(results []KeyValue) error {
	pts, err := allPointers(results)
	if err != nil {
		return err
	}
	sort.Sort(pts)
	keys := pts.Strings()
	header := w.getHeader(pts)

	for i, key := range keys {
		record := toTransposedRecord(results, key, header[i])
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

func allPointers(results []KeyValue) (pointers pointers, err error) {
	set := make(map[string]bool, 0)
	for _, result := range results {
		for _, key := range result.Keys() {
			if !set[key] {
				set[key] = true
				pointer, err := jsonpointer.New(key)
				if err != nil {
					return nil, err
				}
				pointers = append(pointers, pointer)
			}
		}
	}
	return
}

func (w *CSVWriter) getHeader(pointers pointers) []string {
	switch w.HeaderStyle {
	case JSONPointerStyle:
		return pointers.Strings()
	case SlashStyle:
		return pointers.Slashes()
	case DotNotationStyle:
		return pointers.DotNotations(false)
	case DotBracketStyle:
		return pointers.DotNotations(true)
	default:
		return pointers.Strings()
	}
}

func toRecord(kv KeyValue, keys []string) []string {
	record := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := kv[key]; ok {
			record = append(record, toString(value))
		} else {
			record = append(record, "")
		}
	}
	return record
}

func toTransposedRecord(results []KeyValue, key string, header string) []string {
	record := make([]string, 0, len(results)+1)
	record = append(record, header)
	for _, result := range results {
		if value, ok := result[key]; ok {
			record = append(record, toString(value))
		} else {
			record = append(record, "")
		}
	}
	return record
}
