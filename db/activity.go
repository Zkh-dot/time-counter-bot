package db

import (
	"errors"
	"strconv"
	"strings"

	"TimeCounterBot/common"
	"slices"
)

// addActivity добавляет новую активность и возвращает её ID.
func addActivity(activity Activity) (int64, error) {
	result := GormDB.Create(&activity)
	return activity.ID, result.Error
}

// ParseAndAddActivity принимает строку активности в формате
// "Область / Область поуже / ... / Активность" и добавляет её в базу.
func ParseAndAddActivity(userID common.UserID, activityStr string) error {
	parts := strings.Split(activityStr, " / ")
	var parentActivityID int64 = -1

	existingActivities, err := GetSimpleActivities(userID)
	if err != nil {
		return err
	}

	for i, part := range parts {
		isLeaf := false
		if i == len(parts)-1 {
			isLeaf = true
		}

		idx := slices.IndexFunc(existingActivities, func(a Activity) bool {
			return a.Name == part && a.ParentActivityID == parentActivityID && a.IsLeaf == isLeaf
		})

		if idx == -1 {
			newActivity := Activity{
				UserID:           int64(userID),
				Name:             part,
				ParentActivityID: parentActivityID,
				IsLeaf:           isLeaf,
			}
			newID, err := addActivity(newActivity)
			if err != nil {
				return err
			}
			parentActivityID = newID
			existingActivities = append(existingActivities, newActivity)
		} else {
			parentActivityID = existingActivities[idx].ID
		}
	}
	return nil
}

// activityDFS выполняет обход активностей для построения полных путей.
func activityDFS(activities []Activity, vertex int, stack *[]string, ans *[]ActivityRoute) {
	if activities[vertex].IsLeaf {
		*ans = append(*ans, ActivityRoute{
			Name:   strings.Join(*stack, " / ") + " / " + activities[vertex].Name,
			LeafID: activities[vertex].ID,
		})
		return
	}

	*stack = append(*stack, activities[vertex].Name)

	for i, a := range activities {
		if a.ParentActivityID == activities[vertex].ID {
			activityDFS(activities, i, stack, ans)
		}
	}

	*stack = (*stack)[:len(*stack)-1]
}

// buildActivities строит массив ActivityRoute из списка активностей.
func buildActivities(activities []Activity) []ActivityRoute {
	var routes []ActivityRoute
	for i, activity := range activities {
		if activity.ParentActivityID == -1 {
			var stack []string
			activityDFS(activities, i, &stack, &routes)
		}
	}
	return routes
}

// GetFullActivityNameByID возвращает полный путь активности по её ID.
func GetFullActivityNameByID(activityID int64, userID common.UserID) (string, error) {
	routes, err := GetFullActivities(userID)
	if err != nil {
		return "", err
	}
	for _, route := range routes {
		if route.LeafID == activityID {
			return route.Name, nil
		}
	}
	return "", errors.New("Activity not found: " + strconv.FormatInt(activityID, 10))
}

// GetSimpleActivities возвращает список активностей пользователя.
func GetSimpleActivities(userID common.UserID) ([]Activity, error) {
	var activities []Activity
	result := GormDB.Where("user_id = ?", userID).Find(&activities)
	return activities, result.Error
}

// GetFullActivities возвращает полное дерево активностей в виде ActivityRoute.
func GetFullActivities(userID common.UserID) ([]ActivityRoute, error) {
	activities, err := GetSimpleActivities(userID)
	if err != nil {
		return nil, err
	}
	return buildActivities(activities), nil
}
