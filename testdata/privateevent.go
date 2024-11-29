package testdata

type Side int

const (
	SideBuy  Side = 1
	SideSell Side = -1
)

type PrivateChannel string

const (
	PrivateChannelOrder           PrivateChannel = "order"
	PrivateChannelOrderUpdate     PrivateChannel = "order_update"
	PrivateChannelTrade           PrivateChannel = "trade"
	PrivateChannelTradeUpdate     PrivateChannel = "trade_update"
	PrivateChannelTradeFastUpdate PrivateChannel = "trade_fast_update"
	PrivateChannelAccount         PrivateChannel = "account"
	PrivateChannelAccountUpdate   PrivateChannel = "account_update"

	// @group Margin
	PrivateChannelMWalletOrder           PrivateChannel = "mwallet_order"
	PrivateChannelMWalletTrade           PrivateChannel = "mwallet_trade"
	PrivateChannelMWalletTradeFastUpdate PrivateChannel = "mwallet_trade_fast_update"
	PrivateChannelMWalletAccount         PrivateChannel = "mwallet_account"
	PrivateChannelMWalletAveragePrice    PrivateChannel = "mwallet_average_price"
	PrivateChannelBorrowing              PrivateChannel = "borrowing"
	PrivateChannelAdRatio                PrivateChannel = "ad_ratio"
	PrivateChannelPoolQuota              PrivateChannel = "borrowing_pool_quota"
)

// @group Misc
const (
	PrivateChannelAveragePrice   PrivateChannel = "average_price"
	PrivateChannelFavoriteMarket PrivateChannel = "favorite_market"
)
