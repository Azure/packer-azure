package utils

func FindShift(a, b string) int {
	shift := 0
	for shift < len(a) {
		i := 0
		for (i + shift < len(a)) && (i < len(b)) && (a[i + shift] == b[i]) {
			i ++;
		}
		if i + shift == len(a) {
			break;
		}
		shift++;
	}
	return shift;
}

func  Clue(a, b string) string {
	shift := FindShift(a, b);
	return string(a[:shift]) + b
}
