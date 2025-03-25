import * as React from "react"
import { Moon, Sun, Computer } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { setThemePreference } from "@/assets/js/CheckTema"

export function ThemeSwitcher() {
  const [theme, setTheme] = React.useState("system")

  React.useEffect(() => {
    const storedTheme = localStorage.getItem('theme') || "system"
    setTheme(storedTheme)
  }, [])

  function onThemeChange(theme) {
    setTheme(theme)
    setThemePreference(theme)
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8 p-0">
          <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
          <span className="sr-only">Alternar tema</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="animate-in fade-in-50 zoom-in-95 duration-200">
        <DropdownMenuItem 
          onClick={() => onThemeChange("light")} 
          className={theme === "light" ? "bg-accent" : ""}
        >
          <Sun className="mr-2 h-4 w-4" />
          <span>Claro</span>
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => onThemeChange("dark")} 
          className={theme === "dark" ? "bg-accent" : ""}
        >
          <Moon className="mr-2 h-4 w-4" />
          <span>Escuro</span>
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => onThemeChange("system")} 
          className={theme === "system" ? "bg-accent" : ""}
        >
          <Computer className="mr-2 h-4 w-4" />
          <span>Sistema</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}