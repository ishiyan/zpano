/// A single vertex of a Polyline, expressed as (offset, value).
pub const Point = struct {
    /// The number of bars back from the Polyline's time.
    offset: i32,
    /// The value (y) at this vertex.
    value: f64,
};

/// Maximum number of points in a polyline.
pub const max_polyline_points = 256;

/// Holds a time stamp and an ordered, variable-length sequence of points.
pub const Polyline = struct {
    time: i64,
    points: [max_polyline_points]Point = undefined,
    points_len: usize = 0,

    /// Creates a new polyline from a slice of points.
    pub fn new(time: i64, points: []const Point) Polyline {
        var p = Polyline{
            .time = time,
            .points_len = @min(points.len, max_polyline_points),
        };
        @memcpy(p.points[0..p.points_len], points[0..p.points_len]);
        return p;
    }

    /// Creates a new empty polyline with no points.
    pub fn empty(time: i64) Polyline {
        return .{ .time = time };
    }

    /// Indicates whether this polyline has no points.
    pub fn isEmpty(self: Polyline) bool {
        return self.points_len == 0;
    }

    /// Returns the points as a slice.
    pub fn pointsSlice(self: *const Polyline) []const Point {
        return self.points[0..self.points_len];
    }
};
