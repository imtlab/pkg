package excel

/*
	Package excel contains utility functions for working with Excel files.
*/

const (
	KColumnIndicatorBase = 1 + 'Z' - 'A'	//	yields 26
)

/*	Utility functions to convert between 0-based column index and Excel's wacky "column indicator string"
	(A through XFD, corresponding to 0 through 16383, max 16384 columns).  What's wacky about it is,
	unlike straightforward base 26 using A through Z as digits, 'A' does not consistently mean 0.
	'A' means 1 in the most significant place (but only when there is more than one place), but
	it means 0 in all less significant places.  Put another way, when there is more than one letter,
	the most significant place is in base 27 while any lower places are in base 26.  Since this means
	that when transitioning from 2 to 3 digits, the place to the right of the least significant
	place changes base, contortions are needed to accommodate this fuckery.  Hence we have to branch
	to different logic based on range of index, with each branch treating the places differently.
	Leave it to Microsuck to confound things that shoulda-coulda been simple.
*/
func IndexToColumnIndicator(index rune) string {
	xRunes := make([]rune, 0, 3)

	if index < ((KColumnIndicatorBase + 1) * KColumnIndicatorBase) {
		//	1 or 2 letters: [base 27, ]base 26
		denominator := KColumnIndicatorBase
		quotient := index / denominator	//	truncated integer division
		index = index % denominator		//	remainder
		if 0 != quotient {
			//	'A' represents 1 in this place
			xRunes = append(xRunes, 'A' + quotient - 1)
		}
		//	'A' represents 0 in this place
		xRunes = append(xRunes, 'A' + index)
	} else {
		//	3 letters: base 27, base 26, base 26
		/*	Accommodate the change in base of the middle letter by:
			starting with the difference: index - ((KColumnIndicatorBase + 1) * KColumnIndicatorBase),
			and compensate by adding 1 to the quotient: index/(26*26)
		*/
		index -= ((KColumnIndicatorBase + 1) * KColumnIndicatorBase)
		denominator := KColumnIndicatorBase * KColumnIndicatorBase
		quotient := index / denominator	//	truncated integer division
		index = index % denominator		//	remainder
		//	'A' represents 1 in this place (but since we need to add 1 to quotient that cancels the usual - 1)
		xRunes = append(xRunes, 'A' + quotient)

		denominator = KColumnIndicatorBase
		quotient = index / denominator	//	truncated integer division
		index = index % denominator		//	remainder
		//	'A' represents 0 in this place
		xRunes = append(xRunes, 'A' + quotient)
		//	'A' represents 0 in this place
		xRunes = append(xRunes, 'A' + index)
	}

	return string(xRunes)
}

func ColumnIndicatorToIndex(columnIndicator string) (columnIndex rune) {
	for rIndex, rValue := range columnIndicator {
		iValue := rValue - 'A'
		if 0 == rIndex {
			columnIndex = iValue
		} else {
			columnIndex = ((1 + columnIndex) * KColumnIndicatorBase) + iValue
		}
	}

	return
}
