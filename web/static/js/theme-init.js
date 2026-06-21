(function () {
    try {
        var key = "carRentalTheme";
        var mode = window.localStorage.getItem(key);
        var allowed = { light: true, dark: true, system: true };

        if (!allowed[mode]) {
            mode = "system";
        }

        var prefersDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
        var resolved = mode === "dark" || (mode === "system" && prefersDark) ? "dark" : "light";
        var root = document.documentElement;

        root.classList.toggle("dark", resolved === "dark");
        root.dataset.themeMode = mode;
        root.dataset.themeResolved = resolved;
    } catch (error) {
        document.documentElement.dataset.themeMode = "system";
    }
})();
