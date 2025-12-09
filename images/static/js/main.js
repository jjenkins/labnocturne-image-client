// Copy to clipboard functionality
document.addEventListener('DOMContentLoaded', () => {
    // Copy button functionality
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

    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Intersection Observer for fade-in animations
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Observe feature cards
    document.querySelectorAll('.feature-card').forEach((card, index) => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(30px)';
        card.style.transition = `all 0.6s ease ${index * 0.1}s`;
        observer.observe(card);
    });

    // Observe pricing cards
    document.querySelectorAll('.pricing-card').forEach((card, index) => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(30px)';
        card.style.transition = `all 0.6s ease ${index * 0.1}s`;
        observer.observe(card);
    });

    // Add keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Cmd/Ctrl + K to focus on first copy button
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            const firstCopyButton = document.querySelector('.copy-button');
            if (firstCopyButton) {
                firstCopyButton.click();
            }
        }
    });

    // Add parallax effect to hero gradient
    let ticking = false;

    function updateParallax(scrollPos) {
        const hero = document.querySelector('.hero');
        if (hero) {
            const heroGradient = hero.querySelector('::before');
            const offset = scrollPos * 0.5;
            hero.style.transform = `translateY(${offset}px)`;
        }
        ticking = false;
    }

    window.addEventListener('scroll', () => {
        const scrollPos = window.scrollY;
        if (!ticking) {
            window.requestAnimationFrame(() => {
                updateParallax(scrollPos);
            });
            ticking = true;
        }
    });

    // Easter egg: Konami code
    let konamiCode = [];
    const konamiSequence = ['ArrowUp', 'ArrowUp', 'ArrowDown', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'ArrowLeft', 'ArrowRight', 'b', 'a'];

    document.addEventListener('keydown', (e) => {
        konamiCode.push(e.key);
        konamiCode = konamiCode.slice(-10);

        if (konamiCode.join('') === konamiSequence.join('')) {
            document.body.style.animation = 'rainbow 5s linear infinite';
            setTimeout(() => {
                document.body.style.animation = '';
            }, 5000);
        }
    });

    // Add rainbow animation for easter egg
    const style = document.createElement('style');
    style.textContent = `
        @keyframes rainbow {
            0% { filter: hue-rotate(0deg); }
            100% { filter: hue-rotate(360deg); }
        }
    `;
    document.head.appendChild(style);

    // Add cursor trail effect on hero section
    const hero = document.querySelector('.hero');
    if (hero) {
        let trails = [];
        const maxTrails = 15;

        hero.addEventListener('mousemove', (e) => {
            if (trails.length >= maxTrails) {
                const oldTrail = trails.shift();
                oldTrail.remove();
            }

            const trail = document.createElement('div');
            trail.style.position = 'absolute';
            trail.style.width = '4px';
            trail.style.height = '4px';
            trail.style.borderRadius = '50%';
            trail.style.background = 'var(--accent)';
            trail.style.pointerEvents = 'none';
            trail.style.left = e.pageX + 'px';
            trail.style.top = e.pageY + 'px';
            trail.style.opacity = '0.6';
            trail.style.transition = 'all 0.5s ease';
            trail.style.zIndex = '1';

            document.body.appendChild(trail);
            trails.push(trail);

            setTimeout(() => {
                trail.style.opacity = '0';
                trail.style.transform = 'scale(0)';
            }, 50);

            setTimeout(() => {
                trail.remove();
                trails = trails.filter(t => t !== trail);
            }, 500);
        });
    }

    // Add typing effect to terminal cursor
    const cursor = document.querySelector('.terminal-cursor');
    if (cursor) {
        setInterval(() => {
            cursor.style.opacity = cursor.style.opacity === '0' ? '1' : '0';
        }, 530);
    }

    // Log a message for curious developers
    console.log('%c[lab_nocturne]', 'color: #00ff88; font-family: monospace; font-size: 16px; font-weight: bold;');
    console.log('%cWelcome, developer! 👋', 'color: #e0e0e0; font-family: monospace; font-size: 14px;');
    console.log('%cLike what you see? Check out our API:', 'color: #a0a0a0; font-family: monospace; font-size: 12px;');
    console.log('%ccurl https://images.labnocturne.com/key', 'color: #00ff88; font-family: monospace; font-size: 12px; background: #111; padding: 4px 8px; border-radius: 4px;');
    console.log('');
    console.log('%cWe\'re hiring! Email: jobs@labnocturne.com', 'color: #a0a0a0; font-family: monospace; font-size: 11px;');
});
