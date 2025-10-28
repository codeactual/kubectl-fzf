package resources

import "testing"

func TestParseResourceType(t *testing.T) {
	testDatas := []struct {
		resourceName string
		resourceType ResourceType
	}{
		{"", ResourceTypeUnknown},
		{"pods", ResourceTypePod},
		{"pod", ResourceTypePod},
		{"statefulsets", ResourceTypeStatefulSet},
		{"sts", ResourceTypeStatefulSet},
	}

	for _, v := range testDatas {
		r := ParseResourceType(v.resourceName)
		if v.resourceType != r {
			t.Errorf("ParseResourceType(%q) = %v, want %v", v.resourceName, r, v.resourceType)
		}
	}
}

func TestGetResourceType(t *testing.T) {
	testDatas := []struct {
		args         []string
		resourceType ResourceType
	}{
		{[]string{""}, ResourceTypeApiResource},
		{[]string{"pods"}, ResourceTypeApiResource},
		{[]string{"pods", ""}, ResourceTypePod},
	}
	for _, testData := range testDatas {
		parsedType := GetResourceType("get", testData.args)
		if parsedType != testData.resourceType {
			t.Errorf("GetResourceType(%q) = %v, want %v", testData.args, parsedType, testData.resourceType)
		}
	}
}

func TestGetResourceSetFromSliceWithErrors(t *testing.T) {
	testDatas := [][]string{
		{"po", "t", "secrets"},
		{"saa", "pod"},
	}
	for _, testData := range testDatas {
		_, err := GetResourceSetFromSlice(testData)
		if err == nil {
			t.Errorf("GetResourceSetFromSlice(%q) expected error, got nil", testData)
		}
	}
}

func TestGetResourceSetFromSlice(t *testing.T) {
	testDatas := [][]string{
		{"pods", "secrets"},
		{"sa", "pod"},
	}
	for _, testData := range testDatas {
		_, err := GetResourceSetFromSlice(testData)
		if err != nil {
			t.Errorf("GetResourceSetFromSlice(%q) unexpected error: %v", testData, err)
		}
	}
}
