//	Package uuid contains helper functions for working with UUIDs.
package uuid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"github.com/imtlab/pkg/loggers"
)

//\\//	package-scope constants and variables

/*
const (
)
*/

var (
	//	set in init()
	pRegexp	*regexp.Regexp
)

//\\//	type definitions (and attached methods)

//\\//	functions

func init() {
	loggers.Info.Println(`Executing uuid.init()`)

	//	sample "54510a02-6855-11ee-8457-46f5a3bf3389"
	pRegexp = regexp.MustCompile(`^[\da-f]{8}-([\da-f]{4}-){3}[\da-f]{12}$`)
}

func Validate(in string) (out string, err error) {
	if out = strings.TrimSpace(in); 0 == len(out) {
		err = errors.New(`No value`)
	} else if !pRegexp.MatchString(out) {
		err = fmt.Errorf(`"%s" does not match UUID pattern`, out)
	}

	return
}
