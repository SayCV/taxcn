package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/user"
	"strings"
)

var (
	infoHeaders = []string{"统计月份", "累计税前收入", "累计社保", "累计专项扣除", "累计扣除基数", "计税收入", "税率", "速算扣除数", "累计应纳税额", "当月纳税额", "当月税后收入"}
)

const TaxcnSrc = `/TaxcnSrc.json`

type Insurance5X1J struct {
	SocialInsuranceBase      float64 `json:"socialInsuranceBase"`
	HousingFundBase          float64 `json:"housingFundBase"`
	SocialInsuranceCompRatio float64 `json:"socialInsuranceCompRatio"`
	HousingFundCompRatio     float64 `json:"housingFundCompRatio"`
	SocialInsurancePersRatio float64 `json:"socialInsurancePersRatio"`
	HousingFundPersRatio     float64 `json:"housingFundPersRatio"`
}

type SpecialDeduction struct {
	EducationChildren float64 `json:"children"`
	EducationMyself   float64 `json:"myself"`
	MedicalCost       float64 `json:"medicalCost"`
	HousingCost       float64 `json:"housingCost"`
	SupportParents    float64 `json:"supportParents"`
}

type Profile struct {
	SalaryMonthly     float64          `json:"salaryMonthly"`
	YearEndBonusRatio float64          `json:"yearEndBonusRatio"`
	YearEndBonusDate  int64            `json:"yearEndBonusDate"`
	Insurance         Insurance5X1J    `json:"insurance"`
	Deduction         SpecialDeduction `json:"deduction"`
	CalcMonths        int64            `json:"calcMonths"`
}

func NewProfile() *Profile {
	profile := &Profile{}
	//fmt.Println(profile.defaultFileName())
	data, err := ioutil.ReadFile(profile.defaultFileName())
	if err != nil { // Set default values:
		profile.YearEndBonusRatio = 1.78
		profile.SalaryMonthly = 10000
		profile.YearEndBonusDate = 6
		profile.CalcMonths = 6
		profile.Insurance.SocialInsuranceBase = 2500
		profile.Insurance.HousingFundBase = 8000
		profile.Insurance.SocialInsuranceCompRatio = 0.05
		profile.Insurance.HousingFundCompRatio = 0.05
		profile.Insurance.SocialInsurancePersRatio = 0.05
		profile.Insurance.HousingFundPersRatio = 0.05
		profile.Deduction.EducationChildren = 0
		profile.Deduction.EducationMyself = 0
		profile.Deduction.MedicalCost = 0
		profile.Deduction.HousingCost = 1000
		profile.Deduction.SupportParents = 1000

		profile.Save()
	} else {
		json.Unmarshal(data, profile)
	}

	return profile
}

// Save serializes settings using JSON and saves them in ~/.TSrc file.
func (profile *Profile) Save() error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(profile.defaultFileName(), data, 0644)
}

//-----------------------------------------------------------------------------
func (profile *Profile) defaultFileName() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir + TaxcnSrc
}

// main ...
func main() {
	var (
		err             error
		totalIncome     float64 //累计税前收入
		totalInsurance  float64 //累计社保
		totalDeduction  float64 //累计专项扣除
		totalBaseNum    float64 //累计扣除基数
		taxIncome       float64 //计税收入
		taxRate         float64 //税率
		deductionNumber float64 //速算扣除数
		lastTotalTax    float64 //上月累计应纳税额
		totalTax        float64 //累计应纳数额
		tax             float64 //当月纳税额
		finalIncome     float64 //当月税后收入
	)

	profile := NewProfile()

	_, err = fmt.Println(strings.Join(infoHeaders, " | "))
	if err != nil {
		return
	}

	for i := int64(0); i <= profile.CalcMonths; i++ {
		if i == profile.YearEndBonusDate {
			totalIncome += profile.SalaryMonthly * profile.YearEndBonusRatio
		}
		totalIncome += profile.SalaryMonthly
		totalInsurance += getInsurance(profile)
		totalDeduction += getDeduction(profile)
		totalBaseNum += profile.Insurance.SocialInsuranceBase
		taxIncome = totalIncome - totalInsurance - totalDeduction - totalBaseNum
		taxRate, deductionNumber = getTaxRate(taxIncome)
		totalTax = taxIncome*taxRate - deductionNumber
		tax = totalTax - lastTotalTax
		finalIncome = profile.SalaryMonthly - getInsurance(profile) - tax
		lastTotalTax = totalTax
		fmt.Printf("%4d月份 | %-12.f | %-8.f | %-12.f | %-12.f | %-8.f | %-4.2f | %-10.f | %-12.f | %-10.f | %-12.f\n", i+1, totalIncome, totalInsurance, totalDeduction, totalBaseNum, taxIncome, taxRate, deductionNumber, totalTax, tax, finalIncome)
	}
}

func getInsurance(profile *Profile) (value float64) {
	value = profile.Insurance.SocialInsuranceBase*profile.Insurance.SocialInsurancePersRatio + profile.Insurance.HousingFundBase*profile.Insurance.HousingFundPersRatio
	return
}

func getDeduction(profile *Profile) (value float64) {
	value = profile.Deduction.EducationChildren + profile.Deduction.EducationMyself + profile.Deduction.HousingCost + profile.Deduction.MedicalCost + profile.Deduction.SupportParents
	return
}

func getTaxRate(taxIncome float64) (rate float64, num float64) {
	if taxIncome < 36000 {
		rate = 0.03
		num = 0
	} else if taxIncome < 144000 {
		rate = 0.1
		num = 2520
	} else if taxIncome < 300000 {
		rate = 0.2
		num = 16920
	} else if taxIncome < 420000 {
		rate = 0.25
		num = 31920
	} else if taxIncome < 660000 {
		rate = 0.3
		num = 52920
	} else if taxIncome < 960000 {
		rate = 0.35
		num = 85920
	} else {
		rate = 0.45
		num = 181920
	}
	return
}
