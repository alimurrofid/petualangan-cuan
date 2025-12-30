import Swal, { type SweetAlertResult } from 'sweetalert2'

export function useSwal() {
  
  // Enforce high z-index for SweetAlert2 to overlap Shadcn/Radix dialogs
  if (typeof document !== 'undefined') {
    const styleId = 'swal-z-index-fix'
    if (!document.getElementById(styleId)) {
      const style = document.createElement('style')
      style.id = styleId
      style.innerHTML = `
        .swal2-container {
          z-index: 100000 !important;
          pointer-events: auto !important; /* Ensure clicks are captured */
        }
        /* Ensure overlay blocks interactions */
        .swal2-backdrop-show {
          pointer-events: auto !important; 
        }
      `
      document.head.appendChild(style)
    }
  }

  const isDarkMode = () => document.documentElement.classList.contains('dark')

  const fire = (options: any) => {
    return Swal.fire({
      heightAuto: false, // Prevent body shifting/resizing which breaks layout/modals
      returnFocus: false, // Prevent returning focus to element which might trigger other dialogs
      customClass: {
        popup: isDarkMode() ? 'swal2-dark-popup' : '',
        title: isDarkMode() ? 'swal2-dark-title' : '',
        htmlContainer: isDarkMode() ? 'swal2-dark-html' : '',
      },
      background: isDarkMode() ? '#1f2937' : '#ffffff',
      color: isDarkMode() ? '#f3f4f6' : '#1f2937',
      ...options
    })
  }

  const success = (title: string, text: string = '') => {
    return fire({
      icon: 'success',
      title,
      text,
      confirmButtonColor: '#10B981',
    })
  }

  const error = (title: string, text: string = '') => {
    return fire({
      icon: 'error',
      title,
      text,
      confirmButtonColor: '#EF4444',
    })
  }

  const warning = (title: string, text: string = '') => {
    return fire({
      icon: 'warning',
      title,
      text,
      confirmButtonColor: '#F59E0B',
    })
  }
  
  const confirm = async (
    title: string, 
    text: string = 'Anda yakin ingin melakukan ini?', 
    confirmButtonText: string = 'Ya, Lanjutkan!',
    cancelButtonText: string = 'Batal'
  ): Promise<boolean> => {
    const result = await fire({
      title,
      text,
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#3B82F6',
      cancelButtonColor: '#EF4444',
      confirmButtonText,
      cancelButtonText
    }) as SweetAlertResult

    return result.isConfirmed
  }

  const confirmDelete = async (itemType: string): Promise<boolean> => {
    return confirm(
      'Apakah Anda yakin?',
      `Data ${itemType} akan dihapus permanen.`,
      'Ya, Hapus!',
      'Batal'
    )
  }

  // Handle interacting outside of Shadcn Dialog when SweetAlert is open
  const handleSwalInteractOutside = (e: Event) => {
    const target = e.target as HTMLElement;
    if (target?.closest('.swal2-container')) {
      e.preventDefault();
    }
  }

  return {
    fire,
    success,
    error,
    warning,
    confirm,
    confirmDelete,
    handleSwalInteractOutside
  }
}
