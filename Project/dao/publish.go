// Publish 对视频数据的数据库操作
// 创建人：吴润泽
// 创建时间：2022-5-15

package dao

import (
	"Project/models"
	"Project/utils"
	"gorm.io/gorm"
	"time"
)

// CreateVideoByData 将视频的一些链接存储到数据库
// 参数 :
//	title: 视频标题
//  userId : 作者 id
//	playUrl: 视频播放地址
// 	coverUrl: 视频封面地址
// 返回值：
//		Error(nullable)

func CreateVideoByData(title string, authorId int64, playUrl string, coverUrl string) error {
	// 存储相关路径
	id, _ := utils.VideoIDWorker.NextID()
	video := models.Video{
		ID:         id,         // 雪花 ID
		AuthorID:   authorId,   // 作者的 id
		Title:      title,      // 视频标题
		PlayUrl:    playUrl,    // 播放地址
		CoverUrl:   coverUrl,   // 封面地址
		CreateTime: time.Now(), // 获取当前时间插入
	}
	err := DB.Debug().Create(&video).Error
	return err
}

// GetVideoList 获取发布列表
// 参数 :
//	title: 视频标题
//  userId : 作者 id
//	playUrl: 视频播放地址
// 	coverUrl: 视频封面地址
// 返回值：
//		视频列表和错误信息
func GetVideoList(authorId int64, userId int64) ([]models.Video, error) {
	// 查询结果
	var videos []models.Video

	// 查询 follow
	queryFollow := DB.Raw("? UNION ALL ?",
		DB.Select("? as user_id, 1 as is_follow", userId).Table("follow"),                            // 自己不能关注自己
		DB.Select("follow.user_id, 1 as is_follow").Where("follower_id = ?", userId).Table("follow"), // 查找当前用户关注的所有用户
	)
	// 查询点赞
	queryFavorite := DB.Select("video_id, 1 as is_favorite").Where("favorite_id = ?", userId).Table("favorite")

	err := DB.Debug().Table("video").
		// 预加载 User，给 user 表加上 is_follow 字段再查找
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("user.*, is_follow").
				Joins("LEFT JOIN (?) AS fo ON user.user_id = fo.user_id", queryFollow).
				Table("user")
		}).
		// 选择返回的字段
		Select("video.*, is_favorite").
		// 按照创建时间降序排列，即时间最晚的在前面
		Order("video.create_time DESC").
		// 筛选条件，作者的 id 为 authorId 之前的视频
		Where("video.author_id = ? ", authorId).
		// 联结是否点赞
		Joins("LEFT JOIN (?) AS fa ON fa.video_id = video.video_id", queryFavorite).
		Find(&videos).Error
	if err != nil {
		// 数据库查询失败，返回错误信息
		return nil, err
	}
	// 使用 Redis 中的数据更新视频信息
	//err = UpdateVideos(videos[:])
	if err != nil {
		// 如果更新出现问题，返回错误
		return nil, err
	}
	return videos, nil
}
