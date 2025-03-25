import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useConfig } from '@/hooks/useConfig';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Copy, Share2, Check, Clipboard, ExternalLink, Maximize2, Minimize2 } from "lucide-react";
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { ThemeSwitcher } from '@/components/ui/theme-switcher';
import { checkThemePreference } from '@/assets/js/CheckTema';

const CodeBlock = ({ node, inline, className, children, ...props }) => {
  const match = /language-(\w+)/.exec(className || '');
  const code = String(children).replace(/\n$/, '');
  const [copied, setCopied] = useState(false);
  const [hovered, setHovered] = useState(false);
  const isDarkMode = document.documentElement.classList.contains('dark');
  
  const copyCode = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      
      // Reproduz som de feedback ao copiar (opcional)
      const audio = new Audio('data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA/+M4wAAAAAAAAAAAAFhpbmcAAAAPAAAAAwAABPEApaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWl3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d////////////////////////////////////////////AAAAAExhdmM1OC4xMwAAAAAAAAAAAAAAACQDQAAAAAAAAAAE8SrE2zwAAAAAAAAAAAAAAAAA/+MYxAAAAANIAAAAAExBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV/+MYxDsAAANIAAAAAFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV');
      audio.volume = 0.2;
      audio.play().catch(e => {
        // Ignora erro caso o navegador bloqueie autoplay
        console.log("Reprodução de áudio bloqueada");
      });
      
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Erro ao copiar código:', err);
    }
  };

  return !inline && match ? (
    <div 
      className="relative group my-4 rounded-md overflow-hidden transition-all duration-200 ease-in-out"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        transform: hovered ? 'translateY(-2px)' : 'translateY(0)',
        boxShadow: hovered ? '0 10px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)' : '0 4px 6px -1px rgba(0,0,0,0.1)',
        transition: 'transform 0.2s ease, box-shadow 0.2s ease'
      }}
    >
      <div className="absolute right-2 top-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button 
                size="icon" 
                variant="secondary" 
                onClick={copyCode} 
                className={`h-7 w-7 bg-gray-800/70 hover:bg-gray-700 transition-all duration-200 ${copied ? 'scale-110' : ''}`}
              >
                {copied ? 
                  <Check className="h-3.5 w-3.5 text-green-400 animate-in zoom-in-50 duration-200" /> : 
                  <Clipboard className="h-3.5 w-3.5" />
                }
              </Button>
            </TooltipTrigger>
            <TooltipContent side="left" className="animate-in fade-in-50 zoom-in-95 duration-200">
              {copied ? 'Copiado!' : 'Copiar código'}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
      <div className="absolute left-4 top-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <span className="px-2 py-1 rounded-md text-xs font-mono bg-gray-700/50 text-gray-200">
          {match[1].toUpperCase()}
        </span>
      </div>
      <SyntaxHighlighter
        {...props}
        style={isDarkMode ? vscDarkPlus : oneLight}
        language={match[1]}
        PreTag="div"
        className="!rounded-md transition-all duration-300"
      >
        {code}
      </SyntaxHighlighter>
    </div>
  ) : (
    <code {...props} className={`${className} rounded-sm px-1 py-0.5 bg-gray-200 dark:bg-gray-800 transition-colors duration-200`}>
      {children}
    </code>
  );
};

export default function MessagePage() {
  const { hash } = useParams();
  const { config, loading: configLoading } = useConfig();
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tg, setTg] = useState(null);
  const [copied, setCopied] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [animateIn, setAnimateIn] = useState(false);

  useEffect(() => {
    // Verifica preferência de tema salva
    checkThemePreference();
    
    // Inicializa o Telegram WebApp se disponível
    const webApp = window.Telegram?.WebApp;
    if (webApp) {
      webApp.ready();
      webApp.expand();
      setTg(webApp);
    }

    // Só faz a requisição se a configuração estiver carregada
    if (configLoading || !config) {
      return;
    }

    const fetchMessage = async () => {
      try {
        setLoading(true);
        
        const headers = {
          'Content-Type': 'application/json'
        };

        // Adiciona o initData apenas se estivermos em um WebApp do Telegram
        if (webApp?.initData) {
          headers['X-Telegram-Init-Data'] = webApp.initData;
        }

        const baseUrl = config.apiUrl;
        console.log('Tentando acessar API em:', `${baseUrl}/api/messages/${hash}`);
        
        const response = await fetch(`${baseUrl}/api/messages/${hash}`, { 
          headers,
          mode: 'cors'
        });
        
        if (!response.ok) {
          throw new Error(`Erro ao carregar mensagem: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        setMessage(data);

        // Ativa animação de entrada após carregar o conteúdo
        setTimeout(() => {
          setAnimateIn(true);
        }, 100);
      } catch (error) {
        console.error('Erro ao buscar mensagem:', error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };

    fetchMessage();
  }, [hash, config, configLoading]);

  // Aplica o tema do Telegram
  useEffect(() => {
    if (tg) {
      document.documentElement.className = tg.colorScheme;
    }
  }, [tg]);

  const copyFullMessage = async () => {
    try {
      await navigator.clipboard.writeText(message.content);
      setCopied(true);
      
      // Reproduz som de feedback ao copiar
      const audio = new Audio('data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA/+M4wAAAAAAAAAAAAFhpbmcAAAAPAAAAAwAABPEApaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWl3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d////////////////////////////////////////////AAAAAExhdmM1OC4xMwAAAAAAAAAAAAAAACQDQAAAAAAAAAAE8SrE2zwAAAAAAAAAAAAAAAAA/+MYxAAAAANIAAAAAExBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV/+MYxDsAAANIAAAAAFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV');
      audio.volume = 0.2;
      audio.play().catch(e => {
        console.log("Reprodução de áudio bloqueada");
      });
      
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Erro ao copiar mensagem:', err);
    }
  };

  const shareMessage = async () => {
    const url = window.location.href;
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'Resposta do Bot',
          text: message.content,
          url: url
        });
      } catch (err) {
        if (err.name !== 'AbortError') {
          console.error('Erro ao compartilhar:', err);
        }
      }
    } else {
      // Fallback para copiar URL
      try {
        await navigator.clipboard.writeText(url);
        alert('Link copiado para a área de transferência!');
      } catch (err) {
        console.error('Erro ao copiar link:', err);
      }
    }
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
    
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(e => {
        console.log(`Erro ao entrar em modo tela cheia: ${e.message}`);
      });
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen p-4">
        <Card className="w-full max-w-md animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="h-12 w-12 rounded-full border-4 border-t-primary border-r-transparent border-b-transparent border-l-transparent animate-spin"></div>
            <p className="mt-4 text-muted-foreground">Carregando mensagem...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center min-h-screen p-4">
        <Card className="w-full max-w-md border-destructive animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
          <CardHeader>
            <CardTitle className="text-destructive">Erro</CardTitle>
            <CardDescription>Não foi possível carregar a mensagem</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!message) {
    return (
      <div className="flex justify-center items-center min-h-screen p-4">
        <Card className="w-full max-w-md animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
          <CardContent className="p-6">
            <div className="text-center">Mensagem não encontrada</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={`container mx-auto px-4 py-8 ${isFullscreen ? 'fixed inset-0 bg-background z-50 overflow-y-auto' : ''} transition-all duration-300`}>
      <div className="fixed top-4 right-4 z-50">
        <ThemeSwitcher />
      </div>
      
      <Card 
        className={`max-w-4xl mx-auto ${animateIn ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500 ease-out`}
        style={{
          boxShadow: isFullscreen ? 'none' : '0 10px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)'
        }}
      >
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="h-8 w-1 bg-primary rounded-full mr-3 animate-pulse"></div>
              <CardTitle className="text-xl font-medium bg-gradient-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                Resposta do Orbi
              </CardTitle>
            </div>
            <div className="flex space-x-2">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={copyFullMessage}
                      className={`h-8 transition-all duration-200 ${copied ? 'bg-green-100 dark:bg-green-900/20 border-green-300 dark:border-green-700' : ''}`}
                    >
                      {copied ? 
                        <Check className="mr-1 h-4 w-4 text-green-500 animate-in zoom-in-50 duration-200" /> : 
                        <Copy className="mr-1 h-4 w-4" />
                      }
                      {copied ? 'Copiado' : 'Copiar'}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="animate-in fade-in-50 zoom-in-95 duration-200">
                    Copiar todo o conteúdo
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="outline" 
                      size="sm"
                      onClick={shareMessage}
                      className="h-8 transition-all duration-200"
                    >
                      <Share2 className="mr-1 h-4 w-4" />
                      Compartilhar
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="animate-in fade-in-50 zoom-in-95 duration-200">
                    Compartilhar esta resposta
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="ghost" 
                      size="icon"
                      onClick={toggleFullscreen}
                      className="h-8 w-8 transition-all duration-200"
                    >
                      {isFullscreen ? 
                        <Minimize2 className="h-4 w-4" /> : 
                        <Maximize2 className="h-4 w-4" />
                      }
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="animate-in fade-in-50 zoom-in-95 duration-200">
                    {isFullscreen ? 'Sair do modo tela cheia' : 'Modo tela cheia'}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="prose prose-slate dark:prose-invert max-w-none">
            <ReactMarkdown
              components={{
                code: CodeBlock
              }}
            >
              {message.content}
            </ReactMarkdown>
          </div>
        </CardContent>
        <CardFooter className="border-t pt-4 text-xs text-muted-foreground flex justify-between items-center">
          <span>Gerado em: {new Date(message.createdAt).toLocaleString()}</span>
          <a 
            href={`https://t.me/share/url?url=${encodeURIComponent(window.location.href)}&text=${encodeURIComponent('Veja esta resposta do Bot AI!')}`}
            target="_blank" 
            rel="noopener noreferrer"
            className="inline-flex items-center text-xs text-muted-foreground hover:text-primary transition-colors duration-200"
          >
            <ExternalLink className="h-3 w-3 mr-1" />
            Abrir no Telegram
          </a>
        </CardFooter>
      </Card>
    </div>
  );
}