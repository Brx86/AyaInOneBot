package ayafwd

const (
	// TARGET_CHAT_ID — 目标群/频道的 chat_id
	//   - bot 必须在目标群内且有发送消息权限
	TARGET_CHAT_ID = int64(-1001804670613)
)

// SOURCE_CHAT_IDS — 源群/频道的 chat_id 列表
//   - 群组：负数，格式为 -100xxxxxxxxxx（如 -1002234567890）
//   - 频道：同上，负数格式
//   - bot 必须在每个源群内且隐私模式已关闭（@BotFather → Bot Settings → Group Privacy → Turn off）
//
// var SOURCE_CHAT_IDS = []int64{-1001687785734} // 朝色的小窝
var SOURCE_CHAT_IDS = []int64{-1001687785734, -1001889285595}
