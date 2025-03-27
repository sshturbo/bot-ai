import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useConfig } from '@/hooks/useConfig';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { 
  Copy, 
  Share2, 
  Check, 
  Clipboard, 
  ExternalLink, 
  Maximize2, 
  Minimize2, 
  Github, 
  MenuIcon,
  MessageSquarePlus,
  MessageCircle,
  History
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
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
      const audio = new Audio('data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA/+M4wAAAAAAAAAAAAFhpbmcAAAAPAAAAAwAABPEApaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWl3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d////////////////////////////////////////////AAAAAExhdmM1OC4xMwAAAAAAAAAAAAAAACQDQAAAAAAAAAAE8SrE2zwAAAAAAAAAAAAAAAAA/+MYxAAAAANIAAAAAExBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV/+MYxDsAAANIAAAAAFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV');
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
      className="relative group my-4 rounded-md overflow-hidden transition-all duration-200 ease-in-out w-full"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        transform: hovered ? 'translateY(-2px)' : 'translateY(0)',
        boxShadow: hovered ? '0 10px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)' : '0 4px 6px -1px rgba(0,0,0,0.1)',
        transition: 'transform 0.2s ease, box-shadow 0.2s ease'
      }}
    >
      <div className="absolute right-2 top-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-10">
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
      <div className="absolute left-4 top-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-10">
        <span className="px-2 py-1 rounded-md text-xs font-mono bg-gray-700/50 text-gray-200">
          {match[1].toUpperCase()}
        </span>
      </div>
      <div className="overflow-x-auto w-full">
        <SyntaxHighlighter
          {...props}
          style={isDarkMode ? vscDarkPlus : oneLight}
          language={match[1]}
          PreTag="div"
          className="!rounded-md transition-all duration-300"
          customStyle={{
            margin: 0,
            padding: '1rem',
            fontSize: '0.875rem',
          }}
          showLineNumbers={true}
          wrapLines={true}
          wrapLongLines={true}
        >
          {code}
        </SyntaxHighlighter>
      </div>
    </div>
  ) : (
    <code {...props} className={`${className} rounded-sm px-1 py-0.5 bg-gray-200 dark:bg-gray-800 transition-colors duration-200 break-all whitespace-pre-wrap`}>
      {children}
    </code>
  );
};

export default function MessagePage() {
  const { hash } = useParams();
  const navigate = useNavigate();
  const { config, loading: configLoading } = useConfig();
  const [message, setMessage] = useState(null);
  const [chatMessages, setChatMessages] = useState([]); 
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tg, setTg] = useState(null);
  const [copied, setCopied] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [animateIn, setAnimateIn] = useState(false);
  const [messageHistory, setMessageHistory] = useState([]);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isNewChat, setIsNewChat] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [newMessage, setNewMessage] = useState('');
  const [currentChatId, setCurrentChatId] = useState(null);

  // Função para buscar histórico de mensagens
  const fetchMessageHistory = async () => {
    try {
      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      const response = await fetch(`${config.apiUrl}/api/chats`, { 
        headers,
        mode: 'cors'
      });
      
      if (!response.ok) {
        throw new Error('Erro ao carregar histórico');
      }
      
      const data = await response.json();
      setMessageHistory(data);
      return data;
    } catch (error) {
      console.error('Erro ao buscar histórico:', error);
      return [];
    }
  };

  // Função para criar novo chat
  const createNewChat = async () => {
    try {
      setLoading(true);
      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      const response = await fetch(`${config.apiUrl}/api/chat/new`, {
        method: 'POST',
        headers,
        mode: 'cors'
      });

      if (!response.ok) {
        throw new Error('Erro ao criar novo chat');
      }
      
      // Limpa o estado atual
      setMessage(null);
      setChatMessages([]);
      setIsNewChat(true);
      setError(null);
      
      // Busca o histórico atualizado de mensagens
      await fetchMessageHistory();
      
    } catch (error) {
      console.error('Erro ao criar novo chat:', error);
      setError('Erro ao criar novo chat: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  // Função para buscar todas as mensagens de um chat específico
  const fetchChatMessages = async (chatId) => {
    try {
      setLoading(true);
      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      const response = await fetch(`${config.apiUrl}/api/chat/${chatId}`, {
        headers,
        mode: 'cors'
      });

      if (!response.ok) {
        throw new Error('Erro ao carregar mensagens do chat');
      }

      const data = await response.json();
      setChatMessages(data);
      setMessage(data.length > 0 ? data[data.length - 1] : null);
      setCurrentChatId(chatId);
      
      // Ativa animação de entrada após carregar o conteúdo
      setTimeout(() => {
        setAnimateIn(true);
      }, 100);

    } catch (error) {
      console.error('Erro ao buscar mensagens do chat:', error);
      setError(error.message);
      setChatMessages([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Somente busca histórico se o config estiver carregado
    if (config && !configLoading) {
      fetchMessageHistory();
    }
  }, [config, configLoading, tg]);

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

    // Só faz a requisição se a configuração estiver carregada e tivermos um hash
    if (configLoading || !config || !hash) {
      return;
    }

    // Verifica se estamos na rota /message/ (mensagem única) ou /chat/ (histórico completo)
    const isMessagePath = window.location.pathname.includes('/message/');
    
    if (isMessagePath) {
      // Se estamos em /message/:hash, busca apenas a mensagem específica
      fetchSingleMessage(hash);
    } else {
      // Se estamos em /chat/:id, busca todas as mensagens do chat
      fetchChatMessages(hash);
    }
  }, [hash, config, configLoading]);

  // Função para buscar apenas uma mensagem específica pelo hash
  const fetchSingleMessage = async (messageHash) => {
    try {
      setLoading(true);
      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      const response = await fetch(`${config.apiUrl}/api/messages/${messageHash}`, {
        headers,
        mode: 'cors'
      });

      if (!response.ok) {
        throw new Error('Erro ao carregar mensagem');
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

  // Carrega o histórico completo quando o usuário clica em "Ver Histórico"
  const handleShowHistory = async () => {
    // Se já temos mensagens do chat carregadas, apenas alterne a visualização
    if (chatMessages.length > 0) {
      setShowHistory(!showHistory);
      return;
    }

    // Verifica se config está disponível
    if (!config || configLoading) {
      console.error("Configuração não disponível ainda");
      setError("Configuração não carregada. Por favor, recarregue a página.");
      return;
    }

    try {
      setLoading(true);

      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      // Primeiro precisamos descobrir o chat_id a partir do hash da mensagem
      const response = await fetch(`${config.apiUrl}/api/chat/${message.hash}`, {
        headers,
        mode: 'cors'
      });

      if (!response.ok) {
        throw new Error('Erro ao carregar histórico do chat');
      }

      const data = await response.json();
      setChatMessages(data);
      setShowHistory(true);
      
    } catch (error) {
      console.error('Erro ao buscar histórico do chat:', error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const copyFullMessage = async () => {
    try {
      await navigator.clipboard.writeText(message.content);
      setCopied(true);
      
      // Reproduz som de feedback ao copiar
      const audio = new Audio('data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA/+M4wAAAAAAAAAAAAFhpbmcAAAAPAAAAAwAABPEApaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWlpaWl3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d////////////////////////////////////////////AAAAAExhdmM1OC4xMwAAAAAAAAAAAAAAACQDQAAAAAAAAAAE8SrE2zwAAAAAAAAAAAAAAAAA/+MYxAAAAANIAAAAAExBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV/+MYxDsAAANIAAAAAFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV');
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

  const handleNewMessage = async () => {
    try {
      const headers = {
        'Content-Type': 'application/json'
      };

      if (tg?.initData) {
        headers['X-Telegram-Init-Data'] = tg.initData;
      }

      const response = await fetch(`${config.apiUrl}/api/chat/new`, {
        method: 'POST',
        headers,
        mode: 'cors',
        body: JSON.stringify({ content: newMessage })
      });

      if (!response.ok) {
        throw new Error('Erro ao enviar mensagem');
      }

      const data = await response.json();
      setChatMessages([...chatMessages, data]);
      setNewMessage('');
      setIsNewChat(false);
    } catch (error) {
      console.error('Erro ao enviar mensagem:', error);
    }
  };

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar Desktop */}
      <aside className="hidden lg:flex flex-col w-[300px] border-r bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 h-screen sticky top-0">
        <div className="p-4">
          <h2 className="text-xl font-bold bg-gradient-to-r from-primary to-blue-500 bg-clip-text text-transparent mb-4">
            Orbi AI
          </h2>
          <Button 
            onClick={createNewChat} 
            className="w-full justify-start"
            variant="outline"
          >
            <MessageSquarePlus className="mr-2 h-4 w-4" />
            Novo Chat
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto">
          <div className="space-y-2 p-4">
            {messageHistory.map((chat) => (
              <Button
                key={chat.id}
                variant={chat.id === currentChatId ? 'secondary' : 'ghost'}
                className="w-full justify-start"
                onClick={() => {
                  navigate(`/chat/${chat.id}`);
                  setIsSheetOpen(false);
                }}
              >
                <MessageCircle className="mr-2 h-4 w-4" />
                <div className="truncate text-left">
                  {chat.preview_message || "Nova conversa"}
                </div>
              </Button>
            ))}
          </div>
        </div>
      </aside>

      <div className="flex-1">
        {/* Barra de navegação */}
        <nav className="sticky top-0 z-50 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b">
          <div className="container mx-auto px-4 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
                <SheetTrigger asChild>
                  <Button variant="ghost" size="icon" className="lg:hidden">
                    <MenuIcon className="h-5 w-5" />
                  </Button>
                </SheetTrigger>
                <SheetContent side="left" className="w-[300px] sm:w-[400px]">
                  <SheetHeader>
                    <SheetTitle>Conversas</SheetTitle>
                  </SheetHeader>
                  <div className="py-4">
                    <Button 
                      onClick={() => {
                        createNewChat();
                        setIsSheetOpen(false);
                      }} 
                      className="w-full justify-start"
                      variant="outline"
                    >
                      <MessageSquarePlus className="mr-2 h-4 w-4" />
                      Novo Chat
                    </Button>
                  </div>
                  <div className="space-y-2">
                    {loading ? (
                      // Mostrar skeleton loading enquanto carrega
                      Array.from({ length: 3 }).map((_, i) => (
                        <div key={i} className="animate-pulse flex space-x-4 p-2">
                          <div className="h-4 w-4 rounded-full bg-muted"></div>
                          <div className="flex-1 space-y-2">
                            <div className="h-4 bg-muted rounded"></div>
                          </div>
                        </div>
                      ))
                    ) : messageHistory.length > 0 ? (
                      messageHistory.map((chat) => (
                        <Button
                          key={chat.id}
                          variant={chat.id === currentChatId ? 'secondary' : 'ghost'}
                          className="w-full justify-start"
                          onClick={() => {
                            navigate(`/chat/${chat.id}`);
                            setIsSheetOpen(false);
                          }}
                        >
                          <MessageCircle className="mr-2 h-4 w-4" />
                          <div className="truncate text-left">
                            {chat.preview_message || "Nova conversa"}
                          </div>
                        </Button>
                      ))
                    ) : (
                      <div className="text-center py-4 text-muted-foreground">
                        Nenhuma conversa encontrada
                      </div>
                    )}
                  </div>
                </SheetContent>
              </Sheet>
              <h1 className="text-xl font-bold bg-gradient-to-r from-primary to-blue-500 bg-clip-text text-transparent lg:hidden">
                Orbi AI
              </h1>
            </div>
            <div className="flex items-center space-x-4">
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
                  <TooltipContent>
                    {isFullscreen ? 'Sair do modo tela cheia' : 'Modo tela cheia'}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <a
                      href="https://github.com/sshturbo/bot-ai"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center hover:text-primary transition-colors"
                    >
                      <Github className="h-5 w-5" />
                    </a>
                  </TooltipTrigger>
                  <TooltipContent>
                    Ver no GitHub
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <ThemeSwitcher />
            </div>
          </div>
        </nav>

        {/* Conteúdo principal */}
        <main className="container mx-auto px-2 sm:px-4 py-8 w-full max-w-full overflow-hidden">
          {loading ? (
            <div className="flex justify-center items-center min-h-[calc(100vh-6rem)]">
              <Card className="w-full max-w-4xl animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
                <CardContent className="flex flex-col items-center justify-center py-16">
                  <div className="h-12 w-12 rounded-full border-4 border-t-primary border-r-transparent border-b-transparent border-l-transparent animate-spin"></div>
                  <p className="mt-4 text-muted-foreground">Carregando mensagem...</p>
                </CardContent>
              </Card>
            </div>
          ) : error ? (
            <div className="flex justify-center items-center min-h-[calc(100vh-6rem)]">
              <Card className="w-full max-w-4xl border-destructive animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
                <CardHeader>
                  <CardTitle className="text-destructive">Erro</CardTitle>
                  <CardDescription>Não foi possível carregar a mensagem</CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-destructive">
                    {error?.message || error?.toString() || "Ocorreu um erro ao carregar os dados"}
                  </p>
                </CardContent>
              </Card>
            </div>
          ) : isNewChat ? (
            <div className="flex justify-center items-center min-h-[calc(100vh-6rem)]">
              <Card className="w-full max-w-4xl animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
                <CardContent className="p-6">
                  <div className="text-center space-y-6">
                    <div>
                      <h2 className="text-xl font-bold mb-4">Bem-vindo ao Orbi AI!</h2>
                      <p className="text-muted-foreground">Você está em um novo chat.</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          ) : !message ? (
            <div className="flex justify-center items-center min-h-[calc(100vh-6rem)]">
              <Card className="w-full max-w-4xl animate-in fade-in-50 slide-in-from-bottom-8 duration-500">
                <CardContent className="p-6">
                  <div className="text-center">Mensagem não encontrada</div>
                </CardContent>
              </Card>
            </div>
          ) : (
            <div className="space-y-6 overflow-hidden">
              {/* Card principal que mostrará a resposta da IA ou o histórico do chat */}
              <Card 
                className={`w-full sm:max-w-4xl mx-auto ${animateIn ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'} transition-all duration-500 ease-out overflow-hidden`}
                style={{
                  boxShadow: isFullscreen ? 'none' : '0 10px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)'
                }}
              >
                <CardHeader className="pb-2">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div className="flex items-center">
                      <div className="h-8 w-1 bg-primary rounded-full mr-3 animate-pulse"></div>
                      <CardTitle className="text-xl font-medium bg-gradient-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                        {showHistory ? "Histórico do Chat" : "Resposta do Orbi"}
                      </CardTitle>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {!showHistory && (
                        <>
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
                        </>
                      )}
                      
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button 
                              variant="outline" 
                              size="sm"
                              onClick={handleShowHistory}
                              className="h-8 transition-all duration-200"
                            >
                              <History className="mr-1 h-4 w-4" />
                              {showHistory ? "Ver Resposta" : "Ver Histórico"}
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent side="bottom" className="animate-in fade-in-50 zoom-in-95 duration-200">
                            {showHistory ? "Voltar para a resposta" : "Ver o histórico completo do chat"}
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-2 overflow-hidden">
                  {!showHistory ? (
                    // Mostra apenas a resposta da IA selecionada
                    <div className="prose prose-slate dark:prose-invert max-w-none overflow-hidden">
                      <div className="whitespace-pre-wrap break-words">
                        <ReactMarkdown 
                          components={{ 
                            code: CodeBlock,
                            // Força quebra de linha em textos longos
                            p: ({children}) => (
                              <p className="whitespace-pre-wrap break-words">{children}</p>
                            ),
                            // Força quebra de linha em links longos
                            a: ({children, href}) => (
                              <a href={href} className="break-all hover:text-primary transition-colors duration-200">{children}</a>
                            )
                          }}
                        >
                          {message.content}
                        </ReactMarkdown>
                      </div>
                    </div>
                  ) : (
                    // Mostra o histórico completo do chat com mensagens alinhadas em lados diferentes
                    <div className="space-y-6">
                      {chatMessages.map((msg, index) => (
                        <div 
                          key={index} 
                          className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
                        >
                          <div 
                            className={`rounded-lg p-3 sm:p-4 max-w-[95%] sm:max-w-[85%] md:max-w-[80%] ${
                              msg.role === 'user' 
                                ? 'bg-primary/10 border border-primary/20' 
                                : 'bg-muted/50 border border-muted'
                            }`}
                          >
                            <div className="mb-1 text-xs font-medium">
                              {msg.role === 'user' ? 'Você' : 'Orbi AI'}
                            </div>
                            <div className="prose prose-slate dark:prose-invert max-w-none break-words">
                              <div className="whitespace-pre-wrap">
                                <ReactMarkdown 
                                  components={{ 
                                    code: CodeBlock,
                                    p: ({children}) => (
                                      <p className="whitespace-pre-wrap break-words">{children}</p>
                                    ),
                                    a: ({children, href}) => (
                                      <a href={href} className="break-all hover:text-primary transition-colors duration-200">{children}</a>
                                    )
                                  }}
                                >
                                  {msg.content}
                                </ReactMarkdown>
                              </div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
                <CardFooter className="border-t pt-4 text-xs text-muted-foreground flex flex-col sm:flex-row justify-between items-center gap-2">
                  <span>Gerado em: {message.createdAt ? new Date(message.createdAt).toLocaleString() : 'Data não disponível'}</span>
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
          )}
        </main>
      </div>
    </div>
  );
}