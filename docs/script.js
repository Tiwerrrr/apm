document.addEventListener('DOMContentLoaded', () => {
    const donateBtn = document.getElementById('donateBtn');
    const donateMenu = document.getElementById('donateMenu');

    // Toggle dropdown
    donateBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        donateMenu.classList.toggle('show');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!donateMenu.contains(e.target) && e.target !== donateBtn) {
            donateMenu.classList.remove('show');
        }
    });

    // Prevent closing when clicking inside the menu
    donateMenu.addEventListener('click', (e) => {
        e.stopPropagation();
    });
});

// Copy wallet address
function copyWallet() {
    const walletInput = document.getElementById('tonWallet');
    walletInput.select();
    walletInput.setSelectionRange(0, 99999); // For mobile devices
    
    navigator.clipboard.writeText(walletInput.value).then(() => {
        const btn = walletInput.nextElementSibling;
        const originalIcon = btn.innerHTML;
        
        btn.innerHTML = '<i class="fa-solid fa-check" style="color: #27c93f;"></i>';
        
        setTimeout(() => {
            btn.innerHTML = originalIcon;
        }, 2000);
    });
}
