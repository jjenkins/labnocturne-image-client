// Documentation page JavaScript
// Handles HTMX API key fetching, copy buttons, mobile menu, and scroll spy

document.addEventListener('DOMContentLoaded', () => {
    // Initialize all features
    loadStoredAPIKey();
    setupCopyButtons();
    setupMobileNav();
    setupScrollSpy();
});

// API Key Management (HTMX integration)
function handleAPIKeyResponse(event) {
    const detail = event.detail;

    // Check if request was successful
    if (detail.successful && detail.xhr.status === 200) {
        try {
            const data = JSON.parse(detail.xhr.responseText);
            const apiKey = data.api_key;

            if (apiKey) {
                // Store in localStorage
                localStorage.setItem('ln_api_key', apiKey);

                // Populate all examples
                populateAPIKeyExamples(apiKey);

                // Show success feedback
                showAPIKeySuccess(apiKey);
            }
        } catch (err) {
            console.error('Failed to parse API key response:', err);
            showAPIKeyError('Failed to generate API key. Please try again.');
        }
    } else {
        showAPIKeyError('Failed to generate API key. Please try again.');
    }
}

function loadStoredAPIKey() {
    const storedKey = localStorage.getItem('ln_api_key');
    if (storedKey) {
        populateAPIKeyExamples(storedKey);
        showAPIKeySuccess(storedKey);
    }
}

function populateAPIKeyExamples(apiKey) {
    // Find all elements with data-api-key-target attribute
    const elements = document.querySelectorAll('[data-api-key-target="true"]');

    elements.forEach(element => {
        // Replace {{API_KEY}} placeholder in text content
        if (element.textContent) {
            element.innerHTML = element.innerHTML.replace(/\{\{API_KEY\}\}/g, apiKey);
        }

        // Update copy button data attribute if present
        if (element.hasAttribute('data-copy')) {
            const copyText = element.getAttribute('data-copy');
            element.setAttribute('data-copy', copyText.replace(/\{\{API_KEY\}\}/g, apiKey));
        }
    });
}

function showAPIKeySuccess(apiKey) {
    const display = document.getElementById('api-key-display');
    const valueEl = document.getElementById('api-key-value');
    const button = document.getElementById('get-api-key-btn');

    if (display && valueEl) {
        valueEl.textContent = apiKey;
        display.style.display = 'block';
    }

    if (button) {
        button.textContent = 'Key Generated ✓';
        button.disabled = true;
        button.style.opacity = '0.7';
    }
}

function showAPIKeyError(message) {
    const button = document.getElementById('get-api-key-btn');
    if (button) {
        const originalText = button.textContent;
        button.textContent = message;
        button.style.background = 'var(--error)';

        setTimeout(() => {
            button.textContent = originalText;
            button.style.background = '';
        }, 3000);
    }
}

// Copy button functionality (same pattern as homepage)
function setupCopyButtons() {
    const copyButtons = document.querySelectorAll('.copy-button');

    copyButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();
            const textToCopy = button.getAttribute('data-copy');

            try {
                await navigator.clipboard.writeText(textToCopy);

                // Visual feedback
                const originalHTML = button.innerHTML;
                button.innerHTML = `
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                        <path d="M3 8L6 11L13 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                `;

                button.style.transform = 'scale(1.1)';

                setTimeout(() => {
                    button.innerHTML = originalHTML;
                    button.style.transform = '';
                }, 1000);
            } catch (err) {
                console.error('Failed to copy:', err);
            }
        });
    });
}

// Mobile navigation toggle
function setupMobileNav() {
    const toggle = document.querySelector('.mobile-menu-toggle');
    const sidebar = document.querySelector('.docs-sidebar');
    const navLinks = document.querySelectorAll('.docs-nav-link');

    // Create overlay
    const overlay = document.createElement('div');
    overlay.className = 'docs-overlay';
    document.body.appendChild(overlay);

    if (toggle && sidebar) {
        toggle.addEventListener('click', () => {
            toggle.classList.toggle('open');
            sidebar.classList.toggle('open');
            overlay.classList.toggle('show');
        });

        // Close on overlay click
        overlay.addEventListener('click', () => {
            toggle.classList.remove('open');
            sidebar.classList.remove('open');
            overlay.classList.remove('show');
        });

        // Close on nav link click
        navLinks.forEach(link => {
            link.addEventListener('click', () => {
                if (window.innerWidth <= 768) {
                    toggle.classList.remove('open');
                    sidebar.classList.remove('open');
                    overlay.classList.remove('show');
                }
            });
        });
    }
}

// Smooth scroll and active section highlighting
function setupScrollSpy() {
    const navLinks = document.querySelectorAll('.docs-nav-link');
    const sections = document.querySelectorAll('.doc-section');

    // Smooth scroll for anchor links
    navLinks.forEach(link => {
        link.addEventListener('click', function (e) {
            e.preventDefault();
            const targetId = this.getAttribute('href').substring(1);
            const target = document.getElementById(targetId);

            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });

                // Update URL without scrolling
                history.pushState(null, null, `#${targetId}`);
            }
        });
    });

    // Highlight active section on scroll
    const observerOptions = {
        root: null,
        rootMargin: '-20% 0px -70% 0px',
        threshold: 0
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.getAttribute('id');

                // Remove active class from all links
                navLinks.forEach(link => link.classList.remove('active'));

                // Add active class to corresponding link
                const activeLink = document.querySelector(`.docs-nav-link[href="#${id}"]`);
                if (activeLink) {
                    activeLink.classList.add('active');
                }
            }
        });
    }, observerOptions);

    sections.forEach(section => {
        observer.observe(section);
    });

    // Set initial active link based on URL hash
    if (window.location.hash) {
        const targetId = window.location.hash.substring(1);
        const activeLink = document.querySelector(`.docs-nav-link[href="#${targetId}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }
    } else {
        // Default to first link
        const firstLink = document.querySelector('.docs-nav-link');
        if (firstLink) {
            firstLink.classList.add('active');
        }
    }
}

// Make handleAPIKeyResponse available globally for HTMX
window.handleAPIKeyResponse = handleAPIKeyResponse;
