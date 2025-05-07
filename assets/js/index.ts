import 'htmx.org';

(window as any).onScopeValueChange = (event: any) => {
  const value = event.target.value;
  const checkboxes: NodeListOf<HTMLInputElement> = document.querySelectorAll('input[name="scope_checkbox"]:checked');
  const hiddenInput: HTMLInputElement | null = document.getElementById('scope_value') as HTMLInputElement;
  const selectedValues = Array.from(checkboxes).map(checkbox => checkbox.value);
  if (hiddenInput) {
    hiddenInput.value = selectedValues.join(' ');
  }

}

(window as any).parseTokens = function() {
  const hash = window.location.hash.substring(1);
  if (!hash) {
    const result = document.getElementById('result');
    if (result) {
      result.textContent = 'No authentication data found in URL.';
    }
    return {};
  }
  const tokens = {};
  hash.split('&').forEach(part => {
    const [key, value] = part.split('=');
    tokens[key] = decodeURIComponent(value);
  });
  return tokens;
};

