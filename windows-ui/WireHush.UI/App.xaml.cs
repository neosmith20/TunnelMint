using Microsoft.UI.Xaml;

namespace WireHush.UI;

public partial class App : Application
{
    public static MainWindow? MainWindowInstance { get; private set; }

    public App()
    {
        InitializeComponent();
        UnhandledException += OnUnhandledException;
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        MainWindowInstance = new MainWindow();
        MainWindowInstance.Activate();
    }

    private static void OnUnhandledException(object sender, Microsoft.UI.Xaml.UnhandledExceptionEventArgs e)
    {
        // Reference frontend: do not swallow unexpected UI failures.
        // Production logging will be wired to AppLogger before release.
        System.Diagnostics.Debug.WriteLine(e.Exception);
    }
}
