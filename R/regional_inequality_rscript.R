# Install tidyverse from CRAN.
# install.packages("tidyverse")
# The above line needs to be unommented, or run from the terminal. After installed,
# this line can be removed.

# Import tidyverse functions.
library(tidyverse) # for data cleaning and plotting functions.
library(magrittr) # for the %>% pipe operator.

# Read in CSV file.
# As with the python code, we will skip the first line.
region_data <- read.csv(
  "./WID_Data_regions/WID_Data_Metadata/WID_Data_02012024-235418.csv", 
  skip = 1,
  header = TRUE,
  sep = ";"
  )

# Check the head of the data.
head(region_data)

# Check the names of the columns.
# Note that these columns names are slightly different to those read in with
# python.
for (cn in colnames(region_data)) {
  print(cn)
}

# Rename columns
region_data <- region_data %>%
  rename(
    'Africa'= 'sptinc_z_QB.Pre.tax.national.income..Top.10....share.Africa',
    'Asia' = 'sptinc_z_QD.Pre.tax.national.income..Top.10....share.Asia',
    'Latin America' = 'sptinc_z_XL.Pre.tax.national.income..Top.10....share.Latin.America',
    'Europe' = 'sptinc_z_QE.Pre.tax.national.income..Top.10....share.Europe',
    'Middle East' = 'sptinc_z_XM.Pre.tax.national.income..Top.10....share.Middle.East',
    'Oceania' = 'sptinc_z_QF.Pre.tax.national.income..Top.10....share.Oceania',
    'North America' = 'sptinc_z_QP.Pre.tax.national.income..Top.10....share.North.America'
  )

# Check data types.
sapply(region_data, class)

# ggplot2 requires data to be in a 'long' format. This is 'tidy data', wherein
# each row contains only a single value.
# See the article from Hadley Wickham on 'tidy data'
# https://www.jstatsoft.org/article/view/v059i10
# https://cran.r-project.org/web/packages/tidyr/vignettes/tidy-data.html
# Change the data to a 'long' format for later plotting.
# http://www.cookbook-r.com/Manipulating_data/Converting_data_between_wide_and_long_format/
region_data_long <- region_data %>% 
  gather(region, value, 'Africa':'North America', factor_key=TRUE) %>%
head(region_data_long)

# To plot data for each region in its own chart, but have all of those charts collected
# into a single image, we can use facet_grid.
region_data_long %>%
  # first, fitler for data where the values in the 'Year' column are 1980 or greater
  filter(
    1980 <= Year
  ) %>%
  # multiply all values by 100, so that they represent percentages
  transform(value = value * 100) %>%
  # create the basic plot
  ggplot(
    # select the columns to use for each axis
    aes(
      x = Year,
      y = value,
      group = Percentile,
      color = Percentile
    )
  ) + 
  # create the facet grid, where each plot represents a region
  # facet_wrap can be used instead to change the layout of the plots
  facet_grid(region~.) +
  # give the charts a title
  ggtitle('Estimated Share of National Income owned by Percentile Group in Major Regions') +
  # give them a y-label
  ylab('Share of National Income Owned by Group (%)') +
  # set y-axes to always show a scale of 0 to 100
  ylim(0, 100) +
  # plot the lines
  geom_line() +
  # add a stylish theme
  theme_light()
