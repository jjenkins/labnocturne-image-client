// Success Page JavaScript

(function() {
  // Get session_id from URL query parameter
  const urlParams = new URLSearchParams(window.location.search);
  const sessionId = urlParams.get('session_id');

  const loadingState = document.getElementById('loading-state');
  const successState = document.getElementById('success-state');
  const errorState = document.getElementById('error-state');
  const apiKeyText = document.getElementById('api-key-text');
  const planBadge = document.getElementById('plan-badge');
  const errorMessage = document.getElementById('error-message');
  const copyButton = document.getElementById('copy-api-key');

  // If no session_id, show error
  if (!sessionId) {
    showError('No session ID found in URL. Please start the checkout process again.');
    return;
  }

  // Fetch API key from server
  fetchAPIKey(sessionId);

  // Copy API key to clipboard
  if (copyButton) {
    copyButton.addEventListener('click', async function(e) {
      e.preventDefault();
      const apiKey = apiKeyText.textContent;

      try {
        await navigator.clipboard.writeText(apiKey);

        // Update button to show success
        copyButton.classList.add('copied');

        // Change icon to checkmark
        const originalHTML = copyButton.innerHTML;
        copyButton.innerHTML = `
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M3 8L6 11L13 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        `;

        // Reset after 2 seconds
        setTimeout(function() {
          copyButton.innerHTML = originalHTML;
          copyButton.classList.remove('copied');
        }, 2000);
      } catch (err) {
        console.error('Failed to copy:', err);
        alert('Failed to copy API key. Please select and copy manually.');
      }
    });
  }

  async function fetchAPIKey(sessionId) {
    try {
      const response = await fetch('/key/retrieve?session_id=' + encodeURIComponent(sessionId));

      if (!response.ok) {
        const errorData = await response.json();
        const message = errorData.error?.message || 'Failed to retrieve API key';
        showError(message);
        return;
      }

      const data = await response.json();

      // Update UI with API key
      apiKeyText.textContent = data.api_key;
      planBadge.textContent = data.plan + ' plan';

      // Show success state
      showSuccess();
    } catch (error) {
      console.error('Error fetching API key:', error);
      showError('Network error. Please check your connection and try again.');
    }
  }

  function showSuccess() {
    loadingState.style.display = 'none';
    errorState.style.display = 'none';
    successState.style.display = 'block';
  }

  function showError(message) {
    loadingState.style.display = 'none';
    successState.style.display = 'none';
    errorMessage.textContent = message;
    errorState.style.display = 'block';
  }
})();
