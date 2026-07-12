// Client-side PDF generation from already-rendered DOM. The heavy libraries
// (jspdf, html2canvas-pro) are imported dynamically so they only load when an
// admin actually sends something — they stay out of the main bundle.
//
// html2canvas-pro (not classic html2canvas) is required because the app's CSS
// uses oklch colors (Tailwind v4), which the classic library can't parse.

export async function captureElementToPdf(el: HTMLElement): Promise<Blob> {
  const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
    import('html2canvas-pro'),
    import('jspdf'),
  ])

  const canvas = await html2canvas(el, {
    scale: 2, // crisp text/lines when the customer zooms in
    backgroundColor: '#ffffff',
    useCORS: true,
    logging: false,
    // Force the light palette in the screenshot clone so a dark-mode admin
    // still produces a light, customer-friendly document (no on-screen flicker,
    // since this only touches the cloned DOM).
    onclone: (doc: Document) => {
      doc.documentElement.setAttribute('data-theme', 'light')
      doc.documentElement.style.colorScheme = 'light'
    },
  })

  // Fit the capture to A4 width, slicing tall content across pages.
  const pdf = new jsPDF({ unit: 'mm', format: 'a4', orientation: 'portrait' })
  const pageW = pdf.internal.pageSize.getWidth()
  const pageH = pdf.internal.pageSize.getHeight()
  const imgH = (canvas.height * pageW) / canvas.width // full image height at page width

  const img = canvas.toDataURL('image/jpeg', 0.92)
  let remaining = imgH
  let offset = 0
  while (remaining > 0) {
    pdf.addImage(img, 'JPEG', 0, offset === 0 ? 0 : -offset, pageW, imgH)
    remaining -= pageH
    if (remaining > 0) {
      offset += pageH
      pdf.addPage()
    }
  }
  return pdf.output('blob')
}
