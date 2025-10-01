package db

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"TimeCounterBot/common"
	"slices"

	"gopkg.in/yaml.v3"
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

	existingActivities, err := GetSimpleActivities(userID, nil, nil)
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
	routes, err := GetFullActivities(userID, nil)
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
func GetSimpleActivities(userID common.UserID, isMuted *bool, hasMutedLeaves *bool) ([]Activity, error) {
	var activities []Activity
	query := "user_id = ?"
	if isMuted != nil && *isMuted {
		query += " AND is_muted = true"
	} else if isMuted != nil && !*isMuted {
		query += " AND is_muted = false"
	}
	if hasMutedLeaves != nil && *hasMutedLeaves {
		query += " AND has_muted_leaves = true"
	} else if hasMutedLeaves != nil && !*hasMutedLeaves {
		query += " AND has_muted_leaves = false"
	}
	result := GormDB.Where(query, userID).Order("id ASC").Find(&activities)
	return activities, result.Error
}

// GetFullActivities возвращает полное дерево активностей в виде ActivityRoute.
func GetFullActivities(userID common.UserID, isMuted *bool) ([]ActivityRoute, error) {
	activities, err := GetSimpleActivities(userID, isMuted, nil)
	if err != nil {
		return nil, err
	}
	return buildActivities(activities), nil
}

func hasMutedLeaves(activityID int64) (bool, error) {
	var count int64
	err := GormDB.Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id, is_leaf, is_muted FROM activities WHERE id = ?
			UNION ALL
			SELECT a.id, a.is_leaf, a.is_muted
			FROM activities a
			INNER JOIN subtree s ON a.parent_activity_id = s.id
		)
		SELECT COUNT(*) FROM subtree WHERE is_leaf = true AND is_muted = true
	`, activityID).Scan(&count).Error

	return count > 0, err
}

func MuteActivityAndMaybeParents(activityID int64) error {
	// Мьютим саму активность
	if err := GormDB.Model(&Activity{}).
		Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"is_muted":         true,
			"has_muted_leaves": true,
		}).Error; err != nil {
		return err
	}

	// Поднимаемся вверх по дереву
	return muteParentIfNeeded(activityID)
}

func muteParentIfNeeded(childID int64) error {
	var activity Activity
	if err := GormDB.First(&activity, childID).Error; err != nil {
		return err
	}

	if activity.ParentActivityID == -1 {
		return nil
	}
	parentID := activity.ParentActivityID

	// Проверяем, остались ли у родителя незамьюченные дети
	var count int64
	if err := GormDB.Model(&Activity{}).
		Where("parent_activity_id = ? AND is_muted = false", parentID).
		Count(&count).Error; err != nil {
		return err
	}

	hasMuted, err := hasMutedLeaves(parentID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"has_muted_leaves": hasMuted,
	}
	if count == 0 {
		updates["is_muted"] = true
	}

	if err := GormDB.Model(&Activity{}).
		Where("id = ?", parentID).
		Updates(updates).Error; err != nil {
		return err
	}

	if count == 0 {
		return muteParentIfNeeded(parentID)
	}
	return nil
}

func UnmuteActivityAndMaybeParents(activityID int64) error {
	if err := GormDB.Model(&Activity{}).
		Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"is_muted":         false,
			"has_muted_leaves": false,
		}).Error; err != nil {
		return err
	}

	return recursivelyUnmuteParents(activityID)
}

func recursivelyUnmuteParents(childID int64) error {
	var activity Activity
	if err := GormDB.First(&activity, childID).Error; err != nil {
		return err
	}
	if activity.ParentActivityID == -1 {
		return nil
	}
	parentID := activity.ParentActivityID

	// Мы точно знаем, что у родителя есть хотя бы один незамьюченный ребёнок (текущий)
	// Поэтому можно сразу размьютить родителя
	if err := GormDB.Model(&Activity{}).
		Where("id = ?", parentID).
		Update("is_muted", false).Error; err != nil {
		return err
	}

	hasMuted, err := hasMutedLeaves(parentID)
	if err != nil {
		return err
	}

	if err := GormDB.Model(&Activity{}).
		Where("id = ?", parentID).
		Update("has_muted_leaves", hasMuted).Error; err != nil {
		return err
	}

	return recursivelyUnmuteParents(parentID)
}

// buildActivityTree рекурсивно строит дерево активностей для экспорта.
func buildActivityTree(activities []Activity, parentID int64) []ActivityNode {
	var children []ActivityNode

	for _, activity := range activities {
		if activity.ParentActivityID == parentID {
			node := ActivityNode{
				Name:     activity.Name,
				IsMuted:  activity.IsMuted,
				Children: buildActivityTree(activities, activity.ID),
			}
			children = append(children, node)
		}
	}

	return children
}

// ExportActivitiesToYAML экспортирует все активности пользователя в YAML формат.
func ExportActivitiesToYAML(userID common.UserID) ([]byte, error) {
	activities, err := GetSimpleActivities(userID, nil, nil)
	if err != nil {
		return nil, err
	}

	export := ActivityExport{
		Version:    "1.0",
		UserID:     int64(userID),
		ExportDate: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Activities: buildActivityTree(activities, -1),
	}

	return yaml.Marshal(export)
}

// parseActivityNode рекурсивно парсит узел активности и добавляет его в базу.
func parseActivityNode(node ActivityNode, userID common.UserID, parentID int64) error {
	// Проверяем, существует ли уже такая активность
	activities, err := GetSimpleActivities(userID, nil, nil)
	if err != nil {
		return err
	}

	// Определяем, является ли узел листом (нет дочерних элементов)
	isLeaf := len(node.Children) == 0

	// Ищем существующую активность
	idx := slices.IndexFunc(activities, func(a Activity) bool {
		return a.Name == node.Name && a.ParentActivityID == parentID && a.IsLeaf == isLeaf
	})

	var activityID int64
	if idx == -1 {
		// Создаем новую активность
		newActivity := Activity{
			UserID:           int64(userID),
			Name:             node.Name,
			ParentActivityID: parentID,
			IsLeaf:           isLeaf,
			IsMuted:          node.IsMuted,
		}
		activityID, err = addActivity(newActivity)
		if err != nil {
			return err
		}
	} else {
		// Обновляем существующую активность
		activityID = activities[idx].ID
		err = GormDB.Model(&Activity{}).
			Where("id = ?", activityID).
			Update("is_muted", node.IsMuted).Error
		if err != nil {
			return err
		}
	}

	// Обрабатываем дочерние узлы
	for _, child := range node.Children {
		if err := parseActivityNode(child, userID, activityID); err != nil {
			return err
		}
	}

	return nil
}

// ImportActivitiesFromYAML импортирует активности из YAML данных.
func ImportActivitiesFromYAML(data []byte, userID common.UserID) error {
	var export ActivityExport
	if err := yaml.Unmarshal(data, &export); err != nil {
		return err
	}

	// Проверяем версию формата
	if export.Version != "1.0" {
		return errors.New("unsupported export format version: " + export.Version)
	}

	// Импортируем каждый корневой узел
	for _, activity := range export.Activities {
		if err := parseActivityNode(activity, userID, -1); err != nil {
			return err
		}
	}

	return nil
}

// DeleteActivityRecursive удаляет активность и все её дочерние активности.
func DeleteActivityRecursive(activityID int64, userID common.UserID) error {
	// Сначала получаем все активности пользователя
	activities, err := GetSimpleActivities(userID, nil, nil)
	if err != nil {
		return err
	}

	// Находим активность для удаления
	var targetActivity *Activity
	for _, activity := range activities {
		if activity.ID == activityID {
			targetActivity = &activity
			break
		}
	}

	if targetActivity == nil {
		return errors.New("activity not found")
	}

	// Рекурсивно удаляем все дочерние активности
	if err := deleteChildrenRecursive(activityID, activities); err != nil {
		return err
	}

	// Удаляем саму активность
	if err := GormDB.Delete(&Activity{}, activityID).Error; err != nil {
		return err
	}

	// Удаляем все связанные логи активности
	if err := GormDB.Where("activity_id = ?", activityID).Delete(&ActivityLog{}).Error; err != nil {
		return err
	}

	return nil
}

// deleteChildrenRecursive рекурсивно удаляет все дочерние активности.
func deleteChildrenRecursive(parentID int64, activities []Activity) error {
	for _, activity := range activities {
		if activity.ParentActivityID == parentID {
			// Рекурсивно удаляем детей этой активности
			if err := deleteChildrenRecursive(activity.ID, activities); err != nil {
				return err
			}

			// Удаляем активность
			if err := GormDB.Delete(&Activity{}, activity.ID).Error; err != nil {
				return err
			}

			// Удаляем логи активности
			if err := GormDB.Where("activity_id = ?", activity.ID).Delete(&ActivityLog{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
