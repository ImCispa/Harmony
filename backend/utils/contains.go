package utils

func Contains(slice []int, num int) bool {
    for _, v := range slice {
        if v == num {
            return true
        }
    }
    return false
}