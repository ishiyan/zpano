# Primitives

## Welfords algorithm

https://natural-blogarithm.com/post/variance-welford-vs-numpy/

The idea of Welford’s algorithm is to decompose the calculations of the mean and the variance into a form that can easily be updated whenever a new observation comes in.
For the mean this transformation is produced by simply rewriting the formula for $\bar{x}_n$ as the weighted average between the old mean $\bar{x}_{n-1}$ and the new observation $x_n$:

$$\bar{x}_n = \frac{\sum_{i=1}^n x_i}{n}  = \frac{\sum_{i=1}^{n-1} x_i}{n} + \frac{x_n}{n} = \frac{(n-1)\bar{x}_{n-1} + x_n}{n} = \bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n}$$

The decomposition for the variance is slightly more complicated. Only looking at the enumerator (the sum of squares of differences) of the formula for $s^2_n$ from above we can derive it in the following way:

$$\begin{aligned}
&\sum_{i=1}^n (x_i - \bar{x}_n)^2  =  \sum_{i=1}^n (x_i - \underbrace{(\bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n})}_{\text{decomposition of }\bar{x}_n})^2 = \sum_{i=1}^n ((x_i - \bar{x}_{n-1}) - (\frac{x_n - \bar{x}_{n-1}}{n}))^2\\\\
=&\sum_{i=1}^n (x_i - \bar{x}_{n-1})^2 - \underbrace{2 \sum_{i=1}^n (x_i - \bar{x}_{n-1})(\frac{x_n - \bar{x}_{n-1}}{n})}_{=2 \frac{(x_n - \bar{x}_{n-1})}{n}\sum_{i=1}^n(x_i - \bar{x}_{n-1})} + \underbrace{\sum_{i=1}^n (\frac{x_n - \bar{x}_{n-1}}{n})^2}_{=n \cdot (\frac{x_n - \bar{x}_{n-1}}{n})^2 = \frac{(x_n - \bar{x}_{n-1})^2}{n}} \\\\
=&\underbrace{\sum_{i=1}^{n-1} (x_i - \bar{x}_{n-1})^2}_{=(n-2) \cdot s_{n-1}^2} + \underbrace{(x_n - \bar{x}_{n-1})^2 +  \frac{(x_n - \bar{x}_{n-1})^2}{n}}_{=\frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n}} - 2 \frac{(x_n - \bar{x}_{n-1})}{n}\underbrace{\sum_{i=1}^n(x_i - \bar{x}_{n-1})}_{=(n \cdot \bar{x}_n - n\cdot\bar{x}_{n-1})} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n} - 2 \frac{(x_n - \bar{x}_{n-1})}{n}\underbrace{(n \cdot \bar{x}_n - n\cdot\bar{x}_{n-1})}_{=n \cdot (\bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n} - \bar{x}_{n-1})} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n} - 2 \frac{(x_n - \bar{x}_{n-1})^2}{n} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n-1)(x_n - \bar{x}_{n-1})^2}{n}
\end{aligned}$$
