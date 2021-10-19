package utils

import "testing"

type Persion struct {
	Name string
	Age  int
}

type Student struct {
	Name []string
}

func TestDeepCopy(t *testing.T) {
	p := &Persion{
		Name: "123",
		Age:  18,
	}

	var p1 = &Persion{}

	if err := DeepCopy(p1, p); err != nil {
		t.Logf("Test DeepCopy Fail: %s", err)
		t.Fail()
	}

	if p1.Name != p.Name {
		t.Fatalf("p1 Name %s != p Name: %s", p1.Name, p.Name)
	}

	if p1.Age != p.Age {
		t.Fatalf("p1 Age %d != p Age: %d", p1.Age, p.Age)
	}
}

func TestDeepCopyList(t *testing.T) {
	p := &Student{
		Name: []string{"1", "2"},
	}

	var p1 = &Student{}

	p.Name[0] = "01"
	p.Name[1] = "02"

	if err := DeepCopy(p1, p); err != nil {
		t.Logf("Test DeepCopyList Fail: %s", err)
		t.Fatal()
	}

	if p1.Name[0] == p.Name[0] {
		t.Logf("p1 Name[0]: %s != p Name[0]: %s", p1.Name[0], p.Name[0])
	}

	if p1.Name[1] == p.Name[1] {
		t.Logf("p1 Name[1]: %s != p Name[1]: %s", p1.Name[1], p.Name[1])
	}
}

func TestDeepCopyInterface(t *testing.T) {
	var destV = new(interface{})

	var v1 = 123
	var vv1 interface{} = v1
	err := DeepCopy(destV, &vv1)
	t.Log(*destV, err)
}

func TestDuplicate(t *testing.T) {
	v := map[string]interface{}{"aaa": 123}
	destV := Duplicate(v)

	t.Log(destV)
}
