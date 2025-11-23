document.addEventListener('alpine:init', () => {
    Alpine.data('shortenerApp', () => ({
        url: '',
        alias: '',
        showAlias: false,
        loading: false,
        result: null,
        error: null,

        async submit() {
            this.loading = true;
            this.error = null;
            this.result = null;

            try {
                const response = await fetch('/shorten', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        url: this.url, 
                        alias: this.alias 
                    })
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.error || 'Erro desconhecido');
                }

                this.result = data;
                this.url = '';
                this.alias = '';
                this.showAlias = false;

            } catch (err) {
                this.error = err.message;
            } finally {
                this.loading = false;
            }
        },

        copyLink() {
            if (!this.result) return;
            navigator.clipboard.writeText(this.result.short_url);
            alert("Link copiado!");
        }
    }));
});