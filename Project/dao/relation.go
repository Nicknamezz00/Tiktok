// relation 包，该包包含了关注的数据库操作
// 创建人：龚江炜
// 创建时间：2022-5-25

package dao

import (
	"Project/models"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// RelationAction 关注操作
// 参数 :
//		userId : 请求的用户 id
// 		toUserId : 关注的用户 id
//		actionType : 操作类型： 1 -> 关注， 2 -> 取消关注
// 返回值：
//		如果操作成功，返回 nil， 否则返回错误信息
func RelationAction(userId, toUserId, actionType int64) error {
	if userId == toUserId { // 如果是自己对自己操作，不管
		return errors.New("you cannot operate on yourself")
	}
	var count int64 // 查看有没有对应的 user-follower 对
	_ = DB.Debug().Table("follow").Where("user_id = ? AND follower_id = ?", toUserId, userId).Count(&count).Error
	fmt.Println("=====>", count, " and userID : ", userId, " and toUserId:", toUserId)
	if count+actionType == 2 {
		// count 只有 1 和 0
		// 如果 count = 0,那么不能删除(actionType != 2)
		// 如果 count = 1，那么不可以继续插入(actionType != 1)
		return errors.New("action error")
	}
	userKey := "user_followCount_" + strconv.FormatInt(userId, 10)
	toUesrKey := "user_followerCount_" + strconv.FormatInt(toUserId, 10)
	switch actionType {
	case PUBLISH:
		// 如果是关注操作，插入即可
		follow := models.Follow{
			AuthorID:   userId,
			UserID:     toUserId,
			CreateTime: time.Now(),
		}
		// 插入关注记录
		err := DB.Debug().Create(&follow).Error
		if err != nil {
			// 插入失败，返回错误
			return err
		}
		// 当前用户的关注人数 + 1
		err = IncreaseValue(userKey, models.User{ID: userId}, "follow_count", "user")
		if err != nil {
			// 如果更新数据失败，返回错误
			return err
		}
		// 被关注者的粉丝 + 1
		err = IncreaseValue(toUesrKey, models.User{ID: toUserId}, "follower_count", "user")
		return err
	case DELETE:
		// 删除关注记录
		err := DB.Debug().
			Delete(models.Follow{}, "user_id = ? and follower_id = ?", toUserId, userId).Error
		if err != nil {
			// 插入失败，返回错误
			return err
		}
		// 当前用户的关注人数 - 1
		err = DecreaseValue(userKey, models.User{ID: userId}, "follow_count", "user")
		if err != nil {
			// 如果更新数据失败，返回错误
			return err
		}
		// 被关注者的粉丝 - 1
		err = DecreaseValue(toUesrKey, models.User{ID: toUserId}, "follower_count", "user")
		return err
	default:
		// 防御性
		return errors.New("invalid operation")
	}
}

// GetFollowList 获取关注者列表
// 参数 :
//		queryId : 要查看的用户的 id
//		userId : 当前用户的 id
// 返回值：
//		返回关注者列表和错误信息
func GetFollowList(queryId, userId int64) ([]models.User, error) {
	var users []models.User // 结果
	// 查找当前用户关注的所有用户
	queryUserFollow := DB.Raw("(?) UNION ALL (?)",
		DB.Raw("SELECT ? as user_id, 1 as is_follow", userId), // 自己不能关注自己
		DB.Select("follow.user_id, 1 as is_follow").
			Where("follower_id = ?", userId).Table("follow"), // 查找当前用户关注的所有用户
	)
	// 查找 queryID 关注的用户
	queryQueryFollow := DB.Debug().Table("follow").
		Select("follow.user_id, is_follow").
		Where("follower_id = ?", queryId).
		// 联结当前用户关注的人，完成 is_follow 查询
		Joins("LEFT JOIN (?) AS fo ON fo.user_id = follow.user_id", queryUserFollow)
	// 查询用户信息
	err := DB.Debug().Table("user").
		Select("user.*, is_follow").
		// 右联结 queryId 关注的用户
		Joins("RIGHT JOIN (?) AS query ON user.user_id = query.user_id", queryQueryFollow).
		Find(&users).Error
	if err != nil {
		// 查询失败，返回错误信息
		return nil, err
	}
	// 更新用户信息
	err = UpdateUsers(users[:])
	return users, err
}

// GetFollowerUserList 查询粉丝信息，返回粉丝列表
// 参数 :
//		userId : 用户的 id
//		queryId : 要查询用户的 id
// 返回值：
//		返回根据 queryId 查询出的粉丝列表
func GetFollowerUserList(userId int64, queryId int64) ([]models.User, error) {
	var users []models.User // 结果

	// 查找当前用户关注的所有用户
	queryUserFollow := DB.Raw("(?) UNION ALL (?)",
		DB.Raw("SELECT ? as user_id, 1 as is_follow", userId), // 自己不能关注自己
		DB.Select("follow.user_id, 1 as is_follow").
			Where("follower_id = ?", userId).
			Table("follow"), // 查找当前用户关注的所有用户
	)
	// 查找 queryID 的粉丝
	queryQueryFollow := DB.Debug().Table("follow").
		Select("follow.follower_id AS user_id, is_follow").
		Where("follow.user_id = ?", queryId).
		// 联结当前用户关注的人，完成 is_follow 查询
		Joins("LEFT JOIN (?) AS fo ON fo.user_id = follow.follower_id", queryUserFollow)
	// 查询用户信息
	err := DB.Debug().Table("user").
		Select("user.*, is_follow").
		// 右联结 queryId 的粉丝
		Joins("RIGHT JOIN (?) AS query ON user.user_id = query.user_id", queryQueryFollow).
		Find(&users).Error
	if err != nil {
		// 查询失败
		return nil, err
	}
	// 更新用户信息
	err = UpdateUsers(users[:])
	return users, err
}
