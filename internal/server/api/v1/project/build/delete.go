package build

// Delete
// @Summary 	删除项目
// @Description 删除项目
// @Tags 		项目
// @Accept		application/json
// @Produce		application/json
// @Param		project path string true "项目名称"
// @Param		build path string true "构建名称"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/project/{project}/build/{build} [delete]
