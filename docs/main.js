// ===============================
// ANALYTICS SETUP
// ===============================
const analytics = {
    events: [],
    track: function (name, data = {}) {
        const event = { name, data, time: new Date().toISOString() };
        this.events.push(event);
        console.log('[Analytics]', name, data);
    }
};

analytics.track('page_view', { page: window.location.pathname });

// ===============================
// GOATCOUNTER EVENT TRACKING
// ===============================
function trackEvent(eventName, title) {
    if (window.goatcounter && window.goatcounter.count) {
        window.goatcounter.count({
            path: eventName,
            title: title || eventName,
            event: true
        });
    }
}

// ===============================
// COPY COMMAND
// ===============================
function copyInstallCommand() {
    const command = "brew tap Bharath-code/tap && brew install git-scope";
    
    // Fallback for non-secure contexts
    if (!navigator.clipboard) {
        const ta = document.createElement('textarea');
        ta.value = command;
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        showCopyFeedback();
        return;
    }

    navigator.clipboard.writeText(command).then(() => {
        showCopyFeedback();
        trackEvent('command-copied', 'Copied install command');
    }).catch(err => {
        console.error('Failed to copy: ', err);
    });
}

function showCopyFeedback() {
    const icon = document.getElementById('copy-icon');
    if (!icon) return;
    
    const originalHTML = icon.innerHTML;
    
    // Change to Checkmark
    icon.innerHTML = '<polyline points="20 6 9 17 4 12"></polyline>';
    icon.style.opacity = '1';
    icon.style.color = '#27C93F'; // Green from theme

    setTimeout(() => {
        icon.innerHTML = originalHTML;
        icon.style.opacity = '0.5';
        icon.style.color = 'currentColor';
    }, 2000);
}

// ===============================
// SESSION SUMMARY
// ===============================
let maxScroll = 0;
let scrollMilestones = { 25: false, 50: false, 75: false, 100: false };
window.addEventListener('scroll', () => {
    const scrollPercent = Math.round((window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100);
    if (scrollPercent > maxScroll) {
        maxScroll = scrollPercent;
        [25, 50, 75, 100].forEach(milestone => {
            if (scrollPercent >= milestone && !scrollMilestones[milestone]) {
                scrollMilestones[milestone] = true;
                trackEvent('scroll-' + milestone + '-percent', 'Scrolled ' + milestone + '% of page');
            }
        });
    }
});

window.addEventListener('beforeunload', () => {
    analytics.track('session_end', {
        maxScroll: maxScroll,
        totalEvents: analytics.events.length
    });
});
