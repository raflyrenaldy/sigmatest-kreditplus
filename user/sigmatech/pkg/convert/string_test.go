package convert

import (
	"testing"
)

func TestToString(t *testing.T) {
	type args struct {
		t any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Given int value, When call ToString, Then return string value",
			args: args{
				t: 1,
			},
			want: "1",
		},
		{
			name: "Given string value, When call ToString, Then return string value",
			args: args{
				t: "1",
			},
			want: "1",
		},
		{
			name: "Given float value, When call ToString, Then return string value",
			args: args{
				t: 1.1,
			},
			want: "1.1",
		},
		{
			name: "Given bool value, When call ToString, Then return string value",
			args: args{
				t: true,
			},
			want: "true",
		},
		{
			name: "Given struct value, When call ToString, Then return string value",
			args: args{
				t: struct {
					Name string
				}{
					Name: "test",
				},
			},
			want: "{test}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.args.t); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcat(t *testing.T) {
	type args struct {
		values []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Given int values, When call Concat, Then return string value",
			args: args{
				values: []interface{}{
					1,
					2,
					3,
				},
			},
			want: "123",
		},
		{
			name: "Given multiple type of values, When call Concat, Then return string value",
			args: args{
				values: []interface{}{
					1,
					"2",
					3.3,
					true,
				},
			},
			want: "123.3true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Concat(tt.args.values...); got != tt.want {
				t.Errorf("Concat() = %v, want %v", got, tt.want)
			}
		})
	}
}
