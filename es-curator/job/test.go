package job

import (
	// "es-curator/global"
	"fmt"
	"regexp"
	"sort"
	"time"
	// "strconv"
)

func Test1() {
	// f := regexp.MustCompile()
}



func Test() {
	// fmt.Println(global.configViperConfig.Get("orgname"))
	// if global.Action.Delete_indices.Filters.Type_age.Source != "" {
	// 	fmt.Println(global.Action.Delete_indices.Filters.Type_age.Direction)
	// }
	// fmt.Println("test:",global.ActionStruct.Actions)
	// for i := range global.ActionStruct.Actions {
	// 	fmt.Println(global.ActionStruct.Actions[i].Filters)
	// 	for j := range global.ActionStruct.Actions[i].Filters {
	// 		fmt.Println("int", global.ActionStruct.Actions[i].Filters[j].Range_From)
	// 	}
	// }
	// // fmt.Println(global.EnvConfig.ES.URL)
	host := "225.67mb"
	number := regexp.MustCompile(`(\d+.\d+)`)
	sizetype := regexp.MustCompile(`([a-zA-Z]+)`)
	numberStrings := number.FindAllString(host, -1)
	sizetypeStrings := sizetype.FindAllString(host, -1)
	fmt.Println(numberStrings[0], sizetypeStrings[0])

	var nums = []string{"a", "b", "c", "d"}

	for i := range nums {
		fmt.Println(nums[len(nums)-i-1])
	}

	var compareList []string
	list1 := []string{"1", "2", "3", "4", "5", "6"}
	list2 := []string{"1", "5", "2"}
	list3 := []string{"4", "3", "2", "5", "7"}
	for list1data := range list1 {
		for list2data := range list2 {
			for list3data := range list3 {
				if list2[list2data] == list1[list1data] && list3[list3data] == list1[list1data] {
					compareList = append(compareList, list1[list1data])
					// fmt.Println("list1:", list1[list1data], "--", "list2:", list2[list2data])
				}
			}

		}
	}
	fmt.Println("com:", compareList)
	// numbers := make([]int, len(numberStrings))
	// for i, numberString := range numberStrings {
	//     number, err := strconv.Atoi(numberString[1])
	//     if err != nil {
	//         panic(err)
	//     }
	//     numbers[i] = number
	// }
	// fmt.Println(numbers)
}



func Testsum() {
	xi := []int{10, 10, 10, 10, 10}
	fmt.Println(Sum(xi...))
}

func Sum(xi ...int) int {
	// fmt.Printf("%T\n", xi)
	total := 0
	for _, v := range xi {
		total += v
		if total > 21 {
			total = total - v
			break
		}
	}
	return total
}

func Sort() {

	strs := []string{"c", "a", "b"}
	sort.Strings(strs)
	fmt.Println("Strings:", strs)

	ints := []int{7, 2, 4}
	sort.Ints(ints)
	fmt.Println("Ints:   ", ints)

	s := sort.IntsAreSorted(ints)
	fmt.Println("Sorted: ", s)
}


func Job1() {
	fmt.Println(time.Now(),"this is job 1")
}

func Job2() {
	fmt.Println(time.Now(),"this is job 2")
}

func Job3() {
	fmt.Println(time.Now(),"this is job 3")
}