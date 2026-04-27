/// Identifies the chart pane an indicator output is drawn on.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Pane {
    /// The primary price pane.
    Price = 1,
    /// A dedicated sub-pane for this indicator.
    Own = 2,
    /// Drawing on the parent indicator's pane.
    OverlayOnParent = 3,
}

impl Pane {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Price => "price",
            Self::Own => "own",
            Self::OverlayOnParent => "overlayOnParent",
        }
    }
}

impl std::fmt::Display for Pane {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
