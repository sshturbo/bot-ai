import { useState, useEffect } from 'react';

export function useConfig() {
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadConfig = async () => {
      try {
        const response = await fetch('/config.json');
        if (!response.ok) {
          throw new Error('Falha ao carregar configuração');
        }
        const data = await response.json();
        setConfig(data);
      } catch (err) {
        setError(err.message);
        console.error('Erro ao carregar configuração:', err);
      } finally {
        setLoading(false);
      }
    };

    loadConfig();
  }, []);

  return { config, loading, error };
}