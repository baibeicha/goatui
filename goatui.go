package goatui

import (
	"github.com/baibeicha/goatui/pkg/animation"
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/media"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/security"
	"github.com/baibeicha/goatui/pkg/services/explorer"
	"github.com/baibeicha/goatui/pkg/services/network"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/theme"
	"github.com/baibeicha/goatui/pkg/ui"
	"github.com/baibeicha/goatui/pkg/validation"
	"github.com/baibeicha/goatui/pkg/widgets"
	"github.com/baibeicha/goatui/pkg/window"
)

// Primary Core Types
type (
	Cell      = cell.Cell
	Modifier  = cell.Modifier
	Color     = cell.Color
	ColorType = cell.ColorType
	Buffer    = buffer.Buffer
	Rect      = buffer.Rect
	Style     = style.Style
	Border    = style.Border

	// Alignment Types
	Alignment         = buffer.Alignment
	VerticalAlignment = buffer.VerticalAlignment

	// Table & Widget Types
	TableColumn       = widgets.TableColumn
	TableBorder       = widgets.TableBorder
	VirtualTable      = widgets.VirtualTable
	VirtualList       = widgets.VirtualList
	ItemRenderer      = widgets.ItemRenderer
	TableCellRenderer = widgets.TableCellRenderer
	Tabs              = widgets.Tabs
	TabItem           = widgets.TabItem
	TreeView          = widgets.TreeView
	TreeNode          = widgets.TreeNode
	Checkbox          = widgets.Checkbox
	CheckboxGroup     = widgets.CheckboxGroup
	CheckboxItem      = widgets.CheckboxItem
	CheckboxStyle     = widgets.CheckboxStyle
	RadioStyle        = widgets.RadioStyle
	RadioButton       = widgets.RadioButton
	RadioGroup        = widgets.RadioGroup
	RadioItem         = widgets.RadioItem
	Button            = widgets.Button
	ButtonState       = widgets.ButtonState
	ButtonVariant     = widgets.ButtonVariant
	Divider           = widgets.Divider
	Select            = widgets.Select
	SelectItem        = widgets.SelectItem
	Slider            = widgets.Slider
	Histogram         = widgets.Histogram
	HistogramBar      = widgets.HistogramBar
	Spinner           = widgets.Spinner
	SpinnerType       = widgets.SpinnerType
	TextInput         = widgets.TextInput
	EchoMode          = widgets.EchoMode
	Gauge             = widgets.Gauge
	Sparkline         = widgets.Sparkline
	BrailleCanvas     = widgets.BrailleCanvas

	// Declarative UI & Rich Text Types
	Span         = ui.Span
	Line         = ui.Line
	View         = ui.View
	ViewFunc     = ui.ViewFunc
	LayoutItem   = ui.LayoutItem
	Card         = ui.Card
	StatCard     = ui.StatCard
	Badge        = ui.Badge
	KeyHints     = ui.KeyHints
	KeyHint      = ui.KeyHint
	ToastManager = ui.ToastManager
	ToastItem    = ui.ToastItem
	ToastLevel   = ui.ToastLevel
	RenderStream = ui.RenderStream
	FrameArena   = tea.FrameArena

	// Validation Types
	ValidationResult = validation.Result
	ValidationRule   = validation.Rule
	Validator        = validation.Validator
	FormValidator    = validation.Form
	FormResult       = validation.FormResult
	FormField        = ui.FormField

	// Router Types
	Router       = router.Router
	RouteContext = router.RouteContext
	SessionStore = router.SessionStore
	RedirectMsg  = router.RedirectMsg
	Middleware   = router.Middleware
	HandlerFunc  = router.HandlerFunc

	// Window & Screen Types
	Screen             = window.Screen
	BaseScreen         = window.BaseScreen
	Modal              = window.Modal
	InputModal         = window.InputModal
	Omnibar            = window.Omnibar
	OmniItem           = window.OmniItem
	WindowManager      = window.WindowManager
	NavigateMsg        = window.NavigateMsg
	NavigateReplaceMsg = window.NavigateReplaceMsg
	PopMsg             = window.PopMsg
	ShowModalMsg       = window.ShowModalMsg
	CloseModalMsg      = window.CloseModalMsg
	ToggleOmnibarMsg   = window.ToggleOmnibarMsg
	ToggleRoleMsg      = window.ToggleRoleMsg
	AccessDeniedScreen = window.AccessDeniedScreen

	// Theme Types
	Theme           = theme.Theme
	ColorPalette    = theme.ColorPalette
	BorderConfig    = theme.BorderConfig
	ThemeManager    = theme.ThemeManager
	ThemeChangedMsg = theme.ThemeChangedMsg

	// Security Types
	User            = security.User
	SecurityManager = security.SecurityManager

	// Animation Types
	AnimationController = animation.AnimationController
	AnimationMode       = animation.AnimationMode
	AnimationStatus     = animation.AnimationStatus
	EasingFunc          = animation.EasingFunc
	SpringConfig        = animation.SpringConfig
	SpringSimulation    = animation.SpringSimulation
	SlideDirection      = animation.SlideDirection
	SmoothFloat         = animation.SmoothFloat
	SequenceController  = animation.SequenceController
	ParallelController  = animation.ParallelController
	ParticleSystem      = animation.ParticleSystem
	Particle            = animation.Particle

	// Media Types
	Protocol          = media.Protocol
	ScaleMode         = media.ScaleMode
	ImageWidget       = media.ImageWidget
	VideoPlayerWidget = media.VideoPlayerWidget
	BrailleRenderer   = media.BrailleRenderer
	SixelEncoder      = media.SixelEncoder

	// Service Types
	FileExplorerScreen = explorer.FileExplorerScreen
	FileEntry          = explorer.Entry
	IconMode           = explorer.IconMode
	HTTPViewerScreen   = network.HTTPViewerScreen
	HTTPResponseMsg    = network.HTTPResponseMsg
)

const (
	IconModeASCII    = explorer.IconModeASCII
	IconModeUnicode  = explorer.IconModeUnicode
	IconModeNerdFont = explorer.IconModeNerdFont
	IconModeEmoji    = explorer.IconModeEmoji
)

const (
	AlignLeft   = buffer.AlignLeft
	AlignCenter = buffer.AlignCenter
	AlignRight  = buffer.AlignRight

	AlignTop    = buffer.AlignTop
	AlignMiddle = buffer.AlignMiddle
	AlignBottom = buffer.AlignBottom

	AttrNone          = cell.AttrNone
	AttrBold          = cell.AttrBold
	AttrDim           = cell.AttrDim
	AttrItalic        = cell.AttrItalic
	AttrUnderline     = cell.AttrUnderline
	AttrBlink         = cell.AttrBlink
	AttrReverse       = cell.AttrReverse
	AttrHidden        = cell.AttrHidden
	AttrStrikethrough = cell.AttrStrikethrough
)

// Layout Constraints & Direction
type (
	Direction  = layout.Direction
	Constraint = layout.Constraint
)

const (
	Horizontal = layout.Horizontal
	Vertical   = layout.Vertical
)

// The Elm Architecture Types
type (
	Model         = tea.Model
	Cmd           = tea.Cmd
	Msg           = tea.Msg
	Frame         = tea.Frame
	Program       = tea.Program
	ProgramOption = tea.ProgramOption

	KeyMsg        = tea.KeyMsg
	MouseMsg      = tea.MouseMsg
	HitMsg        = tea.HitMsg
	WindowSizeMsg = tea.WindowSizeMsg
	PasteMsg      = tea.PasteMsg
	FocusMsg      = tea.FocusMsg
	BlurMsg       = tea.BlurMsg
	QuitMsg       = tea.QuitMsg
)

// Driver & Events
type (
	Driver      = driver.Driver
	Key         = input.Key
	KeyType     = input.KeyType
	Mouse       = input.Mouse
	MouseButton = input.MouseButton
	MouseAction = input.MouseAction
)

const (
	// Modifier shortcuts
	ModCtrl  = input.ModCtrl
	ModAlt   = input.ModAlt
	ModShift = input.ModShift

	KeyRune      = input.KeyRune
	KeyEnter     = input.KeyEnter
	KeyEsc       = input.KeyEsc
	KeyBackspace = input.KeyBackspace
	KeyTab       = input.KeyTab
	KeyBacktab   = input.KeyBacktab
	KeySpace     = input.KeySpace
	KeyUp        = input.KeyUp
	KeyDown      = input.KeyDown
	KeyLeft      = input.KeyLeft
	KeyRight     = input.KeyRight
	KeyHome      = input.KeyHome
	KeyEnd       = input.KeyEnd
	KeyPgUp      = input.KeyPgUp
	KeyPgDown    = input.KeyPgDown
	KeyInsert    = input.KeyInsert
	KeyDelete    = input.KeyDelete

	// Mouse Buttons & Actions
	MouseNone       = input.MouseNone
	MouseLeft       = input.MouseLeft
	MouseMiddle     = input.MouseMiddle
	MouseRight      = input.MouseRight
	MouseWheelUp    = input.MouseWheelUp
	MouseWheelDown  = input.MouseWheelDown
	MouseWheelLeft  = input.MouseWheelLeft
	MouseWheelRight = input.MouseWheelRight

	MousePress   = input.MousePress
	MouseRelease = input.MouseRelease
	MouseMotion  = input.MouseMotion
	MouseDrag    = input.MouseDrag

	// Animation Modes
	LoopOnce     = animation.LoopOnce
	LoopRepeat   = animation.LoopRepeat
	LoopPingPong = animation.LoopPingPong

	// Slide Directions
	SlideFromLeft   = animation.SlideFromLeft
	SlideFromRight  = animation.SlideFromRight
	SlideFromTop    = animation.SlideFromTop
	SlideFromBottom = animation.SlideFromBottom

	// Media Protocols & Scale Modes
	ProtoKitty     = media.ProtoKitty
	ProtoITerm2    = media.ProtoITerm2
	ProtoSixel     = media.ProtoSixel
	ProtoHalfBlock = media.ProtoHalfBlock
	ProtoBraille   = media.ProtoBraille

	ScaleFit     = media.ScaleFit
	ScaleFill    = media.ScaleFill
	ScaleStretch = media.ScaleStretch

	// Echo Modes for TextInput
	EchoNormal   = widgets.EchoNormal
	EchoPassword = widgets.EchoPassword
	EchoNone     = widgets.EchoNone

	// Checkbox Styles
	CheckboxBrackets     = widgets.CheckboxBrackets
	CheckboxBox          = widgets.CheckboxBox
	CheckboxCircle       = widgets.CheckboxCircle
	CheckboxCircleFilled = widgets.CheckboxCircleFilled
	CheckboxCross        = widgets.CheckboxCross
	CheckboxCheck        = widgets.CheckboxCheck
	CheckboxDot          = widgets.CheckboxDot
	CheckboxSwitch       = widgets.CheckboxSwitch
	CheckboxToggle       = widgets.CheckboxToggle
	CheckboxCustom       = widgets.CheckboxCustom

	// Radio Styles
	RadioCircle       = widgets.RadioCircle
	RadioCircleFilled = widgets.RadioCircleFilled
	RadioDot          = widgets.RadioDot
	RadioBrackets     = widgets.RadioBrackets
	RadioCustom       = widgets.RadioCustom

	// Button States & Variants
	ButtonNormal         = widgets.ButtonNormal
	ButtonHover          = widgets.ButtonHover
	ButtonPressed        = widgets.ButtonPressed
	ButtonDisabled       = widgets.ButtonDisabled
	ButtonVariantDefault = widgets.ButtonVariantDefault
	ButtonVariantPrimary = widgets.ButtonVariantPrimary
	ButtonVariantSuccess = widgets.ButtonVariantSuccess
	ButtonVariantWarning = widgets.ButtonVariantWarning
	ButtonVariantDanger  = widgets.ButtonVariantDanger
	ButtonVariantGhost   = widgets.ButtonVariantGhost

	// Spinner Types
	SpinnerDots     = widgets.SpinnerDots
	SpinnerLine     = widgets.SpinnerLine
	SpinnerMiniDots = widgets.SpinnerMiniDots
	SpinnerPulse    = widgets.SpinnerPulse
	SpinnerArc      = widgets.SpinnerArc
	SpinnerCircle   = widgets.SpinnerCircle

	// Toast Levels
	ToastInfo    = ui.ToastInfo
	ToastSuccess = ui.ToastSuccess
	ToastWarn    = ui.ToastWarn
	ToastError   = ui.ToastError
)

// Global Constructors and Helpers
var (
	// Program
	NewProgram     = tea.NewProgram
	WithDriver     = tea.WithDriver
	WithCatchCtrlC = tea.WithCatchCtrlC
	Batch          = tea.Batch
	Sequence       = tea.Sequence
	Tick           = tea.Tick
	Quit           = tea.Quit

	// Styling
	NewStyle      = style.NewStyle
	BorderNormal  = style.BorderNormal
	BorderRounded = style.BorderRounded
	BorderThick   = style.BorderThick
	BorderDouble  = style.BorderDouble
	BorderASCII   = style.BorderASCII
	BorderBlock   = style.BorderBlock
	MergeBoxRunes = style.MergeBoxRunes

	// Color Constructors
	ColorHex     = cell.ColorHex
	ColorRGB     = cell.RGB
	Color256     = cell.Color256
	Color16      = cell.Color16
	DefaultColor = cell.DefaultColor

	// Layout Solvers
	Split           = layout.Split
	SplitHorizontal = layout.SplitHorizontal
	SplitVertical   = layout.SplitVertical
	SplitInto       = layout.SplitInto
	Fixed           = layout.Fixed
	Percent         = layout.Percent
	Flex            = layout.Flex
	Min             = layout.Min
	Max             = layout.Max
	Ratio           = layout.Ratio

	// Geometry & Text Metrics
	NewRect     = buffer.NewRect
	NewBuffer   = buffer.NewBuffer
	StringWidth = buffer.StringWidth
	RuneWidth   = buffer.RuneWidth

	// Table & Widget Constructors
	NewTable                   = widgets.NewTable
	NewVirtualTable            = widgets.NewVirtualTable
	NewVirtualList             = widgets.NewVirtualList
	DefaultTextRenderer        = widgets.DefaultTextRenderer
	DefaultAlignedTextRenderer = widgets.DefaultAlignedTextRenderer
	TableBorderRounded         = widgets.TableBorderRounded
	TableBorderNormal          = widgets.TableBorderNormal
	TableBorderClean           = widgets.TableBorderClean
	TableBorderNone            = widgets.TableBorderNone
	NewTabs                    = widgets.NewTabs
	NewTreeView                = widgets.NewTreeView
	NewCheckbox                = widgets.NewCheckbox
	NewCheckboxGroup           = widgets.NewCheckboxGroup
	NewTextInput               = widgets.NewTextInput
	NewGauge                   = widgets.NewGauge
	NewSparkline               = widgets.NewSparkline
	NewBrailleCanvas           = widgets.NewBrailleCanvas
	RenderCellText             = widgets.RenderCellText
	NewRadioButton             = widgets.NewRadioButton
	NewRadioGroup              = widgets.NewRadioGroup
	RadioMarkers               = widgets.RadioMarkers
	CheckboxMarkers            = widgets.CheckboxMarkers
	NewButton                  = widgets.NewButton
	NewHorizontalDivider       = widgets.NewHorizontalDivider
	NewVerticalDivider         = widgets.NewVerticalDivider
	NewSelect                  = widgets.NewSelect
	NewSlider                  = widgets.NewSlider
	NewHistogram               = widgets.NewHistogram
	NewSpinner                 = widgets.NewSpinner

	// Declarative UI & Rich Text Constructors
	NewSpan         = ui.NewSpan
	Text            = ui.Text
	Styled          = ui.Styled
	Bold            = ui.Bold
	Dim             = ui.Dim
	Italic          = ui.Italic
	ColorSpan       = ui.Color
	ColoredSpan     = ui.Colored
	BadgeSpan       = ui.BadgeSpan
	NewLine         = ui.NewLine
	LineFromText    = ui.LineFromText
	VBox            = ui.VBox
	HBox            = ui.HBox
	UIFixed         = ui.Fixed
	UIFlex          = ui.Flex
	UIPercent       = ui.Percent
	UIAuto          = ui.Auto
	UISpacer        = ui.Spacer
	UIPadding       = ui.Padding
	UIPad           = ui.Pad
	UICenter        = ui.Center
	NewCard         = ui.NewCard
	NewStatCard     = ui.NewStatCard
	NewBadge        = ui.NewBadge
	NewKeyHints     = ui.NewKeyHints
	NewToastManager = ui.NewToastManager
	NewRenderStream = ui.NewRenderStream
	NewFrameArena   = tea.NewFrameArena

	// Validation Constructors & Builtin Rules
	NewValidator     = validation.New
	NewForm          = validation.NewForm
	NewFormField     = ui.NewFormField
	ValidationOK     = validation.OK
	ValidationFail   = validation.Fail
	RuleRequired     = validation.Required
	RuleMinLength    = validation.MinLength
	RuleMaxLength    = validation.MaxLength
	RuleLengthRange  = validation.LengthRange
	RuleIntRange     = validation.IntRange
	RuleFloatRange   = validation.FloatRange
	RuleRegex        = validation.Regex
	RuleEmail        = validation.Email
	RuleURL          = validation.URL
	RuleNumeric      = validation.Numeric
	RuleAlpha        = validation.Alpha
	RuleAlphanumeric = validation.Alphanumeric
	RuleCustom       = validation.Custom
	RuleOptional     = validation.Optional

	// Router Constructors & Helpers
	NewRouter       = router.NewRouter
	NewRouteContext = router.NewRouteContext
	Redirect        = router.Redirect

	// Window & Screen Constructors & Helpers
	NewWindowManager      = window.NewWindowManager
	Navigate              = window.Navigate
	NavigateReplace       = window.NavigateReplace
	Pop                   = window.Pop
	NewAlertModal         = window.AlertModal
	AlertModal            = window.AlertModal
	NewConfirmModal       = window.ConfirmModal
	ConfirmModal          = window.ConfirmModal
	NewInputModal         = window.NewInputModal
	NewPasswordModal      = window.NewPasswordModal
	NewOmnibar            = window.NewOmnibar
	ToggleOmnibar         = window.ToggleOmnibar
	NewAccessDeniedScreen = window.NewAccessDeniedScreen
	ToggleRole            = window.ToggleRole

	// Theme Constructors & Presets
	NewThemeManager = theme.NewThemeManager
	DefaultTheme    = theme.Default
	SwitchTheme     = theme.SwitchTheme
	DraculaTheme    = theme.Dracula
	CatppuccinTheme = theme.CatppuccinMocha
	NordTheme       = theme.Nord
	MonokaiTheme    = theme.Monokai
	GoatDarkTheme   = theme.GoatDark
	CyberpunkTheme  = theme.Cyberpunk
	MatrixTheme     = theme.Matrix
	ForestTheme     = theme.Forest
	ParseThemeYAML  = theme.ParseThemeYAML

	// Security Constructors & Guards
	NewSecurityManager = security.NewSecurityManager
	DefaultSecurity    = security.Default
	RequireAuth        = security.RequireAuth
	RequireRole        = security.RequireRole
	RequirePermission  = security.RequirePermission
	RequireSecurity    = security.Require

	// Animation Constructors, Physics, and Interpolation
	NewAnimationController = animation.NewController
	NewSequenceController  = animation.NewSequence
	NewParallelController  = animation.NewParallel
	NewParticleSystem      = animation.NewParticleSystem
	NewSpringSimulation    = animation.NewSpringSimulation
	DefaultSpring          = animation.DefaultSpring
	BouncySpring           = animation.BouncySpring
	GentleSpring           = animation.GentleSpring
	NewSmoothFloat         = animation.NewSmoothFloat
	LerpFloat              = animation.LerpFloat
	LerpInt                = animation.LerpInt
	LerpPoint              = animation.LerpPoint
	LerpRect               = animation.LerpRect
	LerpColor              = animation.LerpColor
	SlideRect              = animation.SlideRect
	FadeColor              = animation.FadeColor
	PulseColor             = animation.PulseColor
	ShimmerOffset          = animation.ShimmerOffset

	// Easing curves
	Linear           = animation.Linear
	EaseInQuad       = animation.EaseInQuad
	EaseOutQuad      = animation.EaseOutQuad
	EaseInOutQuad    = animation.EaseInOutQuad
	EaseInCubic      = animation.EaseInCubic
	EaseOutCubic     = animation.EaseOutCubic
	EaseInOutCubic   = animation.EaseInOutCubic
	EaseInElastic    = animation.EaseInElastic
	EaseOutElastic   = animation.EaseOutElastic
	EaseInOutElastic = animation.EaseInOutElastic
	EaseInBounce     = animation.EaseInBounce
	EaseOutBounce    = animation.EaseOutBounce
	EaseInOutBounce  = animation.EaseInOutBounce

	// Media Constructors & Protocol Detection
	DetectProtocol     = media.DetectProtocol
	CurrentProtocol    = media.CurrentProtocol
	SetMediaProtocol   = media.SetProtocol
	RenderHalfBlock    = media.RenderHalfBlock
	NewBrailleRenderer = media.NewBrailleRenderer
	NewSixelEncoder    = media.NewSixelEncoder
	NewImageWidget     = media.NewImageWidget
	NewImageFromFile   = media.NewImageWidgetFromFile
	NewVideoPlayer     = media.NewVideoPlayer
	NewVideoFromGIF    = media.NewVideoPlayerFromGIF

	// Services: File Explorer & Network
	NewFileExplorerScreen = explorer.NewFileExplorerScreen
	FormatFileSize        = explorer.FormatSize
	NewHTTPViewerScreen   = network.NewHTTPViewerScreen
	FetchURLCmd           = network.FetchURLCmd
)
