using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Windows.Graphics;
using WinRT.Interop;

namespace WireHush.UI;

public sealed partial class MainWindow : Window
{
    private readonly AppWindow _appWindow;

    public MainWindow()
    {
        InitializeComponent();

        Title = "WireHush";
        ExtendsContentIntoTitleBar = true;
        SetTitleBar(TitleBarDragRegion);

        var hwnd = WindowNative.GetWindowHandle(this);
        var windowId = Win32Interop.GetWindowIdFromWindow(hwnd);
        _appWindow = AppWindow.GetFromWindowId(windowId);

        // 1536x1024 is the owner-approved 100% reference canvas. WinUI will
        // continue to use DIPs for content and Windows handles DPI scaling.
        _appWindow.Resize(new SizeInt32(1536, 1024));

        ConfigureTitleBar();
    }

    private void ConfigureTitleBar()
    {
        if (!AppWindowTitleBar.IsCustomizationSupported())
        {
            return;
        }

        var titleBar = _appWindow.TitleBar;
        titleBar.ButtonBackgroundColor = Colors.Transparent;
        titleBar.ButtonInactiveBackgroundColor = Colors.Transparent;
        titleBar.ButtonHoverBackgroundColor = Windows.UI.Color.FromArgb(0x35, 0xFF, 0xFF, 0xFF);
        titleBar.ButtonPressedBackgroundColor = Windows.UI.Color.FromArgb(0x22, 0xFF, 0xFF, 0xFF);
        titleBar.ButtonForegroundColor = Windows.UI.Color.FromArgb(0xFF, 0xF3, 0xF7, 0xFA);
        titleBar.ButtonInactiveForegroundColor = Windows.UI.Color.FromArgb(0xA0, 0xF3, 0xF7, 0xFA);

        ApplyCaptionInsets(titleBar);
        titleBar.LayoutMetricsChanged += (_, _) => ApplyCaptionInsets(titleBar);
    }

    private void ApplyCaptionInsets(AppWindowTitleBar titleBar)
    {
        // Keep Settings / Log / Help out from under native caption buttons.
        HeaderCommands.Margin = new Thickness(0, 0, titleBar.RightInset + 8, 0);
    }
}
