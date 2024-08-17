//	Package utils contains handy commonly used utility functions.
package utils

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"os/exec"
	"path"
	"time"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/constraints"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/imtlab/pkg/loggers"
)

func ComputeProgressDivisor(itemCount int, maxIndicatorCount int) int {
	quotient := itemCount / maxIndicatorCount
	if 0 != itemCount % maxIndicatorCount {
		quotient++
	}
	return quotient
}

func ConvertTimeIsoToUnix(s string) (ts int64, err error) {
	var dt time.Time
	if dt, err = time.Parse(time.RFC3339, s); nil == err {
		ts = dt.Unix()
	}
	return
}

/*
A Duration represents the elapsed time between two instants as an int64 nanosecond count.
type Duration int64

func Since(t Time) Duration		It is shorthand for time.Now().Sub(t).
func Until(t Time) Duration		It is shorthand for t.Sub(time.Now()).

A Time represents an instant in time with nanosecond precision.
Time instants can be compared using the Before, After, and Equal methods.
The Sub method subtracts two instants, producing a Duration.
The Add method adds a Time and a Duration, producing a Time.
*/
func GetSleepDuration(unit time.Duration, lower int, upper int) time.Duration {
	return unit * time.Duration(lower + rand.Intn(1+upper-lower))
}

func Sleep(timeStarted time.Time, durationSleep time.Duration) {
	durationSinceStarted := time.Since(timeStarted)
	if durationSinceStarted < durationSleep {
		//	adjust durationSleep to accommodate the time already spent
		durationSleep -= durationSinceStarted
		time.Sleep(durationSleep)
	}
	//	else don't sleep
}

func XStringFromXInt(xi []int) []string {
	/*	What a drag that we have create a slice of string from the slice of int
		If only there was a strings.Join(a []interface{}, sep string) that would internally call the interface's String() method.
	*/
	//	Create a string slice using strconv.Itoa().  Itoa is shorthand for FormatInt(int64(i), 10).
	xs := make([]string, 0, len(xi))
	for _, value := range xi {
		xs = append(xs, strconv.Itoa(value))
	}
	return xs
}

/*	Returns the string starting at rune having index = runeIndexLow, and rune length no greater than runeLengthLimit.
	Pass runeLengthLimit = 0 for no limit.
	Important: The length returned is merely the number of runes that were extracted from strInput,
	potentially used for decision making by the caller.  Due to the final call to strings.TrimSpace()
	it is not necessarily the rune length of the returned string and certainly not to be construed as
	the byte length of the returned string.
*/
func Substring(strInput string, runeIndexLow uint, runeLengthLimit uint) (strOutput string, runeLengthOutput uint) {
	xRunesInput		:= []rune(strInput)
	runeLengthInput	:= uint(len(xRunesInput))

	if runeIndexLow < runeLengthInput {
		runeIndexHigh := runeIndexLow + runeLengthLimit

		if 0 == runeLengthLimit || runeLengthInput < runeIndexHigh {
			//	the limit will not be applied either because it's 0 or the input string is too short
			runeIndexHigh		= runeLengthInput
			runeLengthOutput	= runeLengthInput - runeIndexLow
		} else {
			//	the input string is long enough that the limit will be applied
			runeLengthOutput	= runeLengthLimit
		}

/*		fmt.Printf("@@@ DEBUG: runeLengthInput = %v; runeLengthOutput = %v\n", runeLengthInput, runeLengthOutput)
		fmt.Printf("@@@ DEBUG: [%v : %v : %v]\n", runeIndexLow, runeIndexHigh, runeIndexHigh)
*/
		/*	Convert back to string from rune slice.
			Slice indexing syntax: [low_index : high_index : optional_max]
				The resulting slice includes the element at low_index, but excludes the element at high_index.
				The third parameter, optional_max, sets the capacity of the resulting slice to (optional_max - low_index).
				By default it will be (cap(xRunesInput) - low_index).
		*/
		strOutput = strings.TrimSpace(string(xRunesInput[runeIndexLow : runeIndexHigh : runeIndexHigh]))
	}

	return
}

//	Returns a string composed of the trailing runeLengthLimit runes from strInput.
func SubstringFromEnd(strInput string, runeLengthLimit uint) (strOutput string) {
	if 0 != runeLengthLimit {
		xRunesInput		:= []rune(strInput)
		runeLengthInput	:= uint(len(xRunesInput))

		if runeLengthInput > runeLengthLimit {
			//	the input string is long enough that the limit will be applied
			runeIndexLow := runeLengthInput - runeLengthLimit
			//	Convert back to string from rune slice.
			strOutput = strings.TrimSpace(string(xRunesInput[runeIndexLow : runeLengthInput]))
		} else {
			strOutput = strInput
		}
	}

	return
}

/*	Homegrown CamelCase function:
I supposed a good camelCase function would be trivially easy to find, possibly even in the standard library.
But I could only find 3rd party packages posted by unknowns, and upon inspection of their code, they all were rather flawed.
Either they operated on the bytes of the string rather than the runes (and thus would fail when non-ASCII characters are used),
or they considered only 4 specific (ASCII) characters as characters to be removed, rather than _any_ punctuation, non-printables,
and a zillion other things that would not be a valid symbolic name in Java.  So I guess I have to write my own.
*/
func isAscii(r rune) bool {
	//	type rune = int32
	//	range [00000000, 00000080) Basic Latin aka ASCII
//	return (r < '\x80')				//	rune(128)
	return (r <= unicode.MaxASCII)	//	'\u007F'
}

func removeDiacritics(in string) (out string, err error) {
	/*
	See	https://stackoverflow.com/questions/26722450/remove-diacritics-using-go
	transform.RemoveFunc is deprecated, it is prefer to use runes.Remove(runes.In(unicode.Mn))
	*/
	/*	from unicode
		unicode.Mn	= The set of Unicode characters in category Mn (Mark, Nonspacing)	https://www.fileformat.info/info/unicode/category/Mn/index.htm
	*/
	/*	from golang.org/x/text/transform
		type Transformer interface {
			Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error)
			Reset()
		}
		func String(t Transformer, s string) (result string, n int, err error)
		func Chain(t ...Transformer) Transformer
	*/
	/*	from golang.org/x/text/runes
		type Set interface {
			Contains(r rune) bool	//	Contains returns true if r is contained in the set.
		}
		type Transformer struct {}	//	Transformer implements the transform.Transformer interface.
		func Remove(s Set) Transformer
	*/
	/*	from golang.org/x/text/unicode/norm
	type Form int
	const (
		NFC Form = iota
		NFD
		NFKC
		NFKD
	)
	func (f Form) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error)
	//	Transform() implements the Transform method of the transform.Transformer interface.

	NFC = Unicode Normalization Form C: Compose
	NFD = Unicode Normalization Form D: Decompose

	Normalization Forms:
	Form                         | Description
	---------------------------- | -----------
	Normalization Form D (NFD)   | Canonical Decomposition
	Normalization Form C (NFC)   | Canonical Decomposition, followed by Canonical Composition
	Normalization Form KD (NFKD) | Compatibility Decomposition
	Normalization Form KC (NFKC) | Compatibility Decomposition, followed by Canonical Composition
	*/
	out, _, err = transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), in)

	return
}//removeDiacritics()

func validForCamelCase(r rune, bFirstChar bool) (bValid bool) {
	if isAscii(r) {
		//	digits are invalid if bFirstChar
		if unicode.IsDigit(r) {
			bValid = !bFirstChar
		} else if unicode.IsLetter(r) {
			bValid = true
		}
	}

	return
}

/*
func TestCamelCase(in string) (out string, err error) {
	var (
		myBool	bool
		myRune	rune
		modifiedRune rune
	)

	myRune = '~'
	myBool = isAscii(myRune)
	fmt.Printf("isAscii('%c') = %v\n", myRune, myBool)

	myRune = '¡'
	myBool = isAscii(myRune)
	fmt.Printf("isAscii('%c') = %v\n", myRune, myBool)
	/ *RESULT
	isAscii('~') = true
	isAscii('¡') = false
	* /

	/ *	look into using  unicode.SimpleFold() to convert say ä to a.
		I doubt this is what I need, but just curious to see what it returns for ä.
	* /
	myRune = 'ä'
	modifiedRune = myRune
	fmt.Printf("modifiedRune = '%c'\n", modifiedRune)
	for {
		modifiedRune = unicode.SimpleFold(modifiedRune)
		fmt.Printf("modifiedRune = '%c'\n", modifiedRune)
		
		if modifiedRune == myRune {
			break
		}
	}
	/ *RESULT
	modifiedRune = 'ä'
	modifiedRune = 'Ä'
	modifiedRune = 'ä'
	* /

	if out, err = removeDiacritics(in); nil != err {
		err = fmt.Errorf(`utils.removeDiacritics() failed: %w`, err)
	}

	return
}//CamelCase()
*/
func CamelCase(in string) (out string, err error) {
	/*	bFirstChar will remain true until the first valid rune is encountered (i.e. must be ASCII letter).
		That rune will be lowercased before being written to the output builder, and bFirstChar will be set to false
		and remain false thereafter.
		bCapitalizeNext will be set to true whenever an invalid rune is discarded AND bFirstChar is false.
		when a valid run is encountered while bCapitalizeNext is true, that rune will be uppercased
	*/
	var (
		bFirstChar		bool = true
		bCapitalizeNext	bool
		builder			strings.Builder
	)

	if in, err = removeDiacritics(in); nil == err {
		for _, r := range in {
			if validForCamelCase(r, bFirstChar) {
				if bFirstChar {
					builder.WriteRune(unicode.ToLower(r))
					bFirstChar = false
				} else {
					if bCapitalizeNext {
						r = unicode.ToUpper(r)
						bCapitalizeNext = false
					}
					builder.WriteRune(r)
				}
			} else {
				if !bFirstChar {
					bCapitalizeNext = true
				}
			}
		}

		out = builder.String()
	} else {
		err = fmt.Errorf(`utils.removeDiacritics() failed: %w`, err)
	}

	return
}//CamelCase()

/*	Determine the number of digits consumed by an unsigned integer.  This is useful for determining
	the number of characters to reserve for such things as an incrementing numeric suffix, etc.
*/
func DigitCount(value uint) (digitCount uint) {
	if 0 == value {
		digitCount = 1
	} else {
		//	Mathematica: 1 + Floor[Log10[value]]
		//func Log10(x float64) float64
		//func Floor(x float64) float64
		//	No need for math.Floor(). Truncation occurs during the conversion from float64 to uint
//		digitCount = 1 + uint(math.Floor(math.Log10(float64(value))))
		digitCount = 1 + uint(math.Log10(float64(value)))
	}

	return
}

func BaseName(fn string, extLimit int) string {
	/*	To guard against goofy filenames like "Mr. Smith Goes to Washington" where a human would say
		there is no extension, but path.Ext() would return ". Smith Goes to Washington" as the extension;
		This method will suppress removing any "extension" longer than extLimit.
	*/
	/*	Ext returns the file name extension used by path.
		The extension is the suffix beginning at the final dot in the final slash-separated element of path;
		it is empty if there is no dot.
	*/
	ext	:= path.Ext(fn)
	extLen := len(ext)
	if (0 == extLen) || (extLimit < extLen) {
		return fn
	}
	return fn[0:len(fn)-extLen]
}

func Extension(fn string, extLimit int) string {
	/*	To guard against goofy filenames like "Mr. Smith Goes to Washington" where a human would say
		there is no extension, but path.Ext() would return ". Smith Goes to Washington" as the extension;
		This method will suppress returning any "extension" longer than extLimit.
	*/
	ext	:= path.Ext(fn)
	extLen := len(ext)
	if (0 == extLen) || (extLimit < extLen) {
		return ""
	}
	return ext
}

func FilenameSplit(filename string, extLimit int) (baseName, extension string) {
	/*	To guard against goofy filenames like "Mr. Smith Goes to Washington" where a human would say
		there is no extension, but path.Ext() would return ". Smith Goes to Washington" as the extension;
		This method will suppress any "extension" longer than extLimit.
	*/
	/*	Ext returns the file name extension used by path.
		The extension is the suffix beginning at the final dot in the final slash-separated element of path;
		it is empty if there is no dot.
	*/
	extension = path.Ext(filename)
	extLen := len(extension)
	if (0 == extLen) || (extLimit < extLen) {
		baseName = filename
		extension = ""
	} else {
		baseName = filename[0:len(filename)-extLen]
	}
	return
}

func ErrorsToMessages(xerr []error) (xMessages []string) {
	xMessages = make([]string, len(xerr))
	for index := range xerr {
		/*	For the standard Go compiler, the internal structure of any string type is declared like:
			type _string struct {
				elements	*byte	//	underlying bytes
				len			int		//	number of bytes
			}
		I'm guessing that when assigning one string variable to another,
		its struct is copied with its elements field still pointing to the source string's byte array.
		As strings are immutable, that array will never be modified.  A new array is allocated
		only when a new value is assigned to a string.  So I expect that the following assignment
		does not cause all the bytes in the string's array to be duplicated in memory.
		*/
		xMessages[index] = xerr[index].Error()
	}

	return
}

func IsExactUint64(f64 float64) (value uint64, exact bool) {
	pBigRat := new(big.Rat)
	//func (z *Rat) SetFloat64(f float64) *Rat
	pBigRat.SetFloat64(f64)

	//func (x *Int) IsUint64() bool
	if pBigRat.Denom().IsUint64() {
		//func (x *Int) Uint64() uint64
		if denom := pBigRat.Denom().Uint64(); 1 == denom {
			exact = true
			value = pBigRat.Num().Uint64()
		}
	}
	//	else leave value = 0, exact = false

	return
}

/*
func EuclidGCD(p, q uint) uint {
	for 0 != q {
		p, q = q, p % q
	}

	return p
}
*/
//	Generic version
/*	Q:	Would this work for negative integers?
	A:	(So far not needed)
*/
func EuclidGCD[T constraints.Unsigned](p, q T) T {
	for 0 != q {
		p, q = q, p % q
	}

	return p
}

func Ceiling[T constraints.Unsigned](dividend, divisor T) (quotient T) {
	quotient = dividend / divisor
	if remainder := dividend % divisor; 0 != remainder {
		quotient++
	}

	return
}

func Plural[T constraints.Integer](count T) (plural string) {
	if 1 != count {
		plural = `s`
	}

	return
}

//\\//	helpers for encoding/csv

/*
func BuildMapStringKeyToIndex(xKeys []string) (mKeyToIndex map[string]int, err error) {
	mKeyToIndex = make(map[string]int)

	for index := range xKeys {
		key := xKeys[index]
		if _, present := mKeyToIndex[key]; present {
			err = fmt.Errorf(`Key "%s" appears more than once`, key)
		} else {
			mKeyToIndex[key] = index
		}
	}

	return
}
*/
//	Generic version
//	See https://tip.golang.org/ref/spec#Type_parameter_declarations
func BuildMapKeyToIndex[T comparable](xKeys []T) (mKeyToIndex map[T]int, err error) {
	mKeyToIndex = make(map[T]int)

	for index := range xKeys {
		key := xKeys[index]
		if _, present := mKeyToIndex[key]; present {
			err = fmt.Errorf(`Key "%v" appears more than once`, key)
		} else {
			mKeyToIndex[key] = index
		}
	}

	return
}

/*	When reading from a file encoded as UTF-8 + BOM, the BOM is read in as part of csvRecord[0],
	so it needs to be removed.
*/
func RemoveBOM(csvRecord []string) {
	xRunes := []rune(csvRecord[0])
	if '\ufeff' == xRunes[0] {
		csvRecord[0] = string(xRunes[1:])
	}
}


/*
func getPaths() {
	//	check existence of source dir (and make sure it's a directory).
	if fileInfoSrc, err := os.Stat(sPathDirSrc); nil == err {
		if !fileInfoSrc.IsDir() {
			fmt.Printf("ERROR: \"%v\" is not a directory\n", sPathDirSrc)
			return
		}
	} else {
		fmt.Println(err)
		return
	}

	//	check existence of destination dir (and make sure it's a directory). create it if it ain't there
	if fileInfoDst, err := os.Stat(sPathDirDst); nil == err {
		if !fileInfoDst.IsDir() {
			fmt.Printf("ERROR: \"%v\" is not a directory\n", fileInfoDst)
			return
		}
	} else {
		if os.IsNotExist(err) {
			//	create the directory
			if err = os.MkdirAll(sPathDirDst, os.ModePerm); nil != err {
				fmt.Println(err)
				return
			}
		}
	}
}
*/

/*	WARNING: This expects that loggers has been Init'ed
	ADDENDUM 2024-07-01:
		No longer important now that the loggers package has an init() that sets its exported
		Info, Warning, and Error pointers to log.Default().
*/
func ExecuteCommand(p *exec.Cmd, bErrIfStderr bool) (stdout string, err error) {
	//\\//	establish the pipes
	//func (c *Cmd) StdoutPipe() (io.ReadCloser, error)
	var rcStdout io.ReadCloser
	if rcStdout, err = p.StdoutPipe(); nil == err {
		//func (c *Cmd) StderrPipe() (io.ReadCloser, error)
		var rcStderr io.ReadCloser
		if rcStderr, err = p.StderrPipe(); nil == err {
			//\\//	start the command
			//func (c *Cmd) Start() error
			if err = p.Start(); nil == err {
				//\\//	slurp STDOUT and STDERR
				//func ReadAll(r Reader) ([]byte, error)
				var xBytesStdout []byte
				if xBytesStdout, err = io.ReadAll(rcStdout); nil == err {
					var xBytesStderr []byte
					if xBytesStderr, err = io.ReadAll(rcStderr); nil == err {
						//\\//	wait for completion
						/*	If the process was started successfully, Wait() will populate p.ProcessState when the command completes.
							ProcessState *os.ProcessState

							//func (p *ProcessState) Success() bool
							Success() reports whether the program exited successfully, such as with exit status 0 on Unix.

							//func (p *ProcessState) ExitCode() int
							ExitCode() returns the exit code of the exited process, or -1 if the process hasn't exited or was terminated by a signal.
						*/
						/*	Since p.Wait() didn't return an error, then p.ProcessState.ExitCode() == 0
							and p.ProcessState.Success() == true.
							Regard any stderr output as indicative of an error, even though the ExitCode is 0.
						*/
						//func (c *Cmd) Wait() error
						if err = p.Wait(); err != nil {
							err = fmt.Errorf(`p.Wait() failed: exit code: %d; error: %w`, p.ProcessState.ExitCode(), err)
						}

						if 0 != len(xBytesStdout) {
							stdout = string(xBytesStdout)
							loggers.Info.Printf(`stdout: %s`, stdout)
						}

						if 0 != len(xBytesStderr) {
							if bErrIfStderr {
								if nil == err {
									err = fmt.Errorf(`stderr: %s`, string(xBytesStderr))
								} else {
									err = fmt.Errorf(`error: %w; stderr: %s`, err, string(xBytesStderr))
								}
							} else {
								loggers.Warning.Printf(`stderr: %s`, string(xBytesStderr))
							}
						}
					} else {
						err = fmt.Errorf(`io.ReadAll(rcStderr) failed: %w`, err)
					}
				} else {
					err = fmt.Errorf(`io.ReadAll(rcStdout) failed: %w`, err)
				}
			} else {
				err = fmt.Errorf(`p.Start() failed: %w`, err)
			}
		} else {
			err = fmt.Errorf(`p.StderrPipe() failed: %w`, err)
		}
	} else {
		err = fmt.Errorf(`p.StdoutPipe(): %w`, err)
	}

	return
}//ExecuteCommand()

