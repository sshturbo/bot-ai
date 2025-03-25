export function setThemePreference(preference) {
    if (preference === 'dark') {
        // Aplica o tema escuro
        document.documentElement.classList.add('dark');
        document.documentElement.classList.remove('light');
        localStorage.setItem('theme', 'dark');
    } else if (preference === 'light') {
        // Aplica o tema claro
        document.documentElement.classList.add('light');
        document.documentElement.classList.remove('dark');
        localStorage.setItem('theme', 'light');
    } else if (preference === 'system') {
        // Remove qualquer tema fixo e verifica o sistema
        localStorage.setItem('theme', 'system');
        applySystemTheme();
    }
}

export function applySystemTheme() {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');

    const applyTheme = (isDark) => {
        if (isDark) {
            document.documentElement.classList.add('dark');
            document.documentElement.classList.remove('light');
        } else {
            document.documentElement.classList.add('light');
            document.documentElement.classList.remove('dark');
        }
    };

    // Aplica o tema atual do sistema
    applyTheme(prefersDark.matches);

    // Listener para mudanças dinâmicas no tema do sistema
    prefersDark.addEventListener('change', (event) => {
        applyTheme(event.matches);
    });
}

export function checkThemePreference() {
    const storedTheme = localStorage.getItem('theme');

    if (storedTheme === 'dark') {
        document.documentElement.classList.add('dark');
        document.documentElement.classList.remove('light');
    } else if (storedTheme === 'light') {
        document.documentElement.classList.add('light');
        document.documentElement.classList.remove('dark');
    } else {
        // Aplica o tema do sistema se estiver configurado para "system" ou não definido
        applySystemTheme();
    }
}
